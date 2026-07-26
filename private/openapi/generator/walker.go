package generator

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"math/rand"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

type GenerationContext struct {
	visited      map[*openapi3.Schema]int
	currentDepth int
	maxDepth     int
	seed         int64
}

func NewGenerationContext(seed int64, maxDepth int) *GenerationContext {
	if maxDepth <= 0 {
		maxDepth = 5
	}
	return &GenerationContext{
		visited:      make(map[*openapi3.Schema]int),
		currentDepth: 0,
		maxDepth:     maxDepth,
		seed:         seed,
	}
}

type Walker struct {
	scalar     *ScalarSynthesizer
	poly       *PolyResolver
	staticSeed bool
	seedSource func() int64
}

const maxGeneratedArrayItems = 100

func NewWalker(staticSeed bool) *Walker {
	return &Walker{
		scalar:     NewScalarSynthesizer(),
		poly:       NewPolyResolver(),
		staticSeed: staticSeed,
		seedSource: rand.Int63,
	}
}

// GenerateFromOperation synthesizes realistic response payload for an OpenAPI Operation definition.
func (w *Walker) GenerateFromOperation(httpMethod, path, operationID string, op *openapi3.Operation, maxDepth int) (int, map[string]string, any, error) {
	if op == nil || op.Responses == nil {
		return 200, nil, map[string]any{}, nil
	}

	statusCode, respRef := selectResponse(op.Responses)

	if respRef == nil || respRef.Value == nil || respRef.Value.Content == nil {
		if respRef == nil || respRef.Value == nil {
			return statusCode, nil, map[string]any{}, nil
		}
	}

	seed := GenerateSeed(httpMethod, path, operationID)
	if !w.staticSeed {
		// Mix operation identity with fresh entropy so unstubbed responses vary
		// between requests while remaining internally coherent for one response.
		seed ^= w.seedSource()
	}
	headers, err := w.generateResponseHeaders(respRef.Value, seed, maxDepth)
	if err != nil {
		return statusCode, nil, nil, err
	}
	if respRef.Value.Content == nil {
		return statusCode, headers, map[string]any{}, nil
	}

	mediaType := respRef.Value.Content.Get("application/json")
	if mediaType == nil || mediaType.Schema == nil || mediaType.Schema.Value == nil {
		// Fallback to any content type if JSON not present
		mediaTypes := make([]string, 0, len(respRef.Value.Content))
		for name := range respRef.Value.Content {
			mediaTypes = append(mediaTypes, name)
		}
		sort.Strings(mediaTypes)
		for _, name := range mediaTypes {
			mt := respRef.Value.Content[name]
			if mt != nil && mt.Schema != nil && mt.Schema.Value != nil {
				mediaType = mt
				break
			}
		}
	}

	if mediaType == nil || mediaType.Schema == nil || mediaType.Schema.Value == nil {
		return statusCode, headers, map[string]any{}, nil
	}

	ctx := NewGenerationContext(seed, maxDepth)

	payload, err := w.generateSchema(ctx, mediaType.Schema.Value)
	if err != nil {
		return statusCode, nil, nil, err
	}

	return statusCode, headers, payload, nil
}

func (w *Walker) generateResponseHeaders(response *openapi3.Response, seed int64, maxDepth int) (map[string]string, error) {
	if response == nil || len(response.Headers) == 0 {
		return nil, nil
	}

	headers := make(map[string]string, len(response.Headers))
	for name, headerRef := range response.Headers {
		if headerRef == nil || headerRef.Value == nil {
			continue
		}
		header := headerRef.Value
		value := header.Example
		if value == nil && header.Schema != nil && header.Schema.Value != nil {
			ctx := NewGenerationContext(seed+int64(GenerateSeed("HEADER", name, "")), maxDepth)
			generated, err := w.generateSchema(ctx, header.Schema.Value)
			if err != nil {
				return nil, fmt.Errorf("generate response header %s: %w", name, err)
			}
			value = generated
		}
		if value == nil {
			continue
		}
		headers[name] = formatHeaderValue(value, header.Explode != nil && *header.Explode)
	}
	return headers, nil
}

func formatHeaderValue(value any, explode bool) string {
	switch value := value.(type) {
	case string:
		return value
	case []any:
		parts := make([]string, len(value))
		for i, item := range value {
			parts[i] = fmt.Sprint(item)
		}
		return strings.Join(parts, ",")
	case map[string]any:
		keys := make([]string, 0, len(value))
		for key := range value {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		parts := make([]string, 0, len(keys)*2)
		for _, key := range keys {
			if explode {
				parts = append(parts, key+"="+fmt.Sprint(value[key]))
			} else {
				parts = append(parts, key, fmt.Sprint(value[key]))
			}
		}
		return strings.Join(parts, ",")
	default:
		if encoded, err := json.Marshal(value); err == nil {
			return string(encoded)
		}
		return fmt.Sprint(value)
	}
}

func (w *Walker) generateSchema(ctx *GenerationContext, schema *openapi3.Schema) (any, error) {
	if schema == nil {
		return nil, nil
	}

	if ctx.currentDepth >= ctx.maxDepth {
		return nil, nil
	}

	ctx.currentDepth++
	defer func() { ctx.currentDepth-- }()

	// Cycle safety check for recursion tracking pointer
	ctx.visited[schema] = ctx.visited[schema] + 1
	defer func() { ctx.visited[schema] = ctx.visited[schema] - 1 }()

	if ctx.visited[schema] > ctx.maxDepth {
		// Break the cycle
		return nil, nil
	}

	// 1. Check allOf
	if len(schema.AllOf) > 0 {
		return w.poly.ResolveAllOf(w, ctx, schema.AllOf)
	}

	// 2. Check oneOf
	if len(schema.OneOf) > 0 {
		return w.poly.ResolveOneOfOrAnyOf(w, ctx, schema.OneOf, schema.Discriminator)
	}

	// 3. Check anyOf
	if len(schema.AnyOf) > 0 {
		return w.poly.ResolveOneOfOrAnyOf(w, ctx, schema.AnyOf, schema.Discriminator)
	}

	// 4. Object Schema
	types := schema.Type.Slice()
	primaryType := ""
	if len(types) > 0 {
		primaryType = types[0]
	}

	if primaryType == "object" || len(schema.Properties) > 0 {
		obj := make(map[string]any)
		for propName, propRef := range schema.Properties {
			if propRef == nil || propRef.Value == nil {
				continue
			}
			propCtx := &GenerationContext{
				visited:      ctx.visited,
				currentDepth: ctx.currentDepth,
				maxDepth:     ctx.maxDepth,
				seed:         ctx.seed + GenerateSeed("PROP", propName, ""),
			}
			val, err := w.generateSchema(propCtx, propRef.Value)
			if err != nil {
				return nil, err
			}
			if val != nil {
				obj[propName] = val
			}
		}
		return obj, nil
	}

	// 5. Array Schema
	if primaryType == "array" || schema.Items != nil {
		if schema.Items == nil || schema.Items.Value == nil {
			return []any{}, nil
		}
		count := 2
		if schema.MinItems > 0 {
			if schema.MinItems > maxGeneratedArrayItems {
				count = maxGeneratedArrayItems
			} else {
				count = int(schema.MinItems)
			}
		}
		if schema.MaxItems != nil && uint64(count) > *schema.MaxItems {
			count = int(*schema.MaxItems)
		}
		items := make([]any, 0, count)
		for i := 0; i < count; i++ {
			itemCtx := &GenerationContext{
				visited:      ctx.visited,
				currentDepth: ctx.currentDepth,
				maxDepth:     ctx.maxDepth,
				seed:         ctx.seed + int64(i),
			}
			itemVal, err := w.generateSchema(itemCtx, schema.Items.Value)
			if err != nil {
				return nil, err
			}
			if itemVal != nil {
				items = append(items, itemVal)
			}
		}
		return items, nil
	}

	// 6. Primitive Leaf Synthesis
	r := rand.New(rand.NewSource(ctx.seed))
	return w.scalar.Synthesize(schema, r), nil
}

func GenerateSeed(httpMethod, path, operationID string) int64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(httpMethod + path + operationID))
	return int64(h.Sum64())
}

func selectResponse(responses *openapi3.Responses) (int, *openapi3.ResponseRef) {
	for _, status := range []int{http.StatusOK, http.StatusCreated} {
		if response := responses.Value(strconv.Itoa(status)); response != nil {
			return status, response
		}
	}
	if response := responses.Default(); response != nil {
		return http.StatusOK, response
	}

	type candidate struct {
		status   int
		key      string
		response *openapi3.ResponseRef
	}
	responseMap := responses.Map()
	candidates := make([]candidate, 0, len(responseMap))
	for key, response := range responseMap {
		status := parseCode(key)
		if status != 0 && response != nil {
			candidates = append(candidates, candidate{status: status, key: key, response: response})
		}
	}
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].status != candidates[j].status {
			return candidates[i].status < candidates[j].status
		}
		return candidates[i].key < candidates[j].key
	})
	if len(candidates) == 0 {
		return http.StatusOK, nil
	}
	return candidates[0].status, candidates[0].response
}

func parseCode(code string) int {
	if len(code) != 3 {
		return 0
	}
	status, err := strconv.Atoi(code)
	if err != nil || status < 100 || status > 599 {
		return 0
	}
	return status
}
