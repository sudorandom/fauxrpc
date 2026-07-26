package generator

import (
	"hash/fnv"
	"math/rand"

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
	scalar *ScalarSynthesizer
	poly   *PolyResolver
}

func NewWalker() *Walker {
	return &Walker{
		scalar: NewScalarSynthesizer(),
		poly:   NewPolyResolver(),
	}
}

// GenerateFromOperation synthesizes realistic response payload for an OpenAPI Operation definition.
func (w *Walker) GenerateFromOperation(httpMethod, path, operationID string, op *openapi3.Operation, maxDepth int) (int, any, error) {
	if op == nil || op.Responses == nil {
		return 200, map[string]any{}, nil
	}

	// Select response code (prefer 200, then 201, then default/first)
	statusCode := 200
	var respRef *openapi3.ResponseRef

	if r := op.Responses.Value("200"); r != nil {
		statusCode = 200
		respRef = r
	} else if r := op.Responses.Value("201"); r != nil {
		statusCode = 201
		respRef = r
	} else if r := op.Responses.Default(); r != nil {
		statusCode = 200
		respRef = r
	} else {
		// Pick first available response status code
		for code, ref := range op.Responses.Map() {
			if ref != nil {
				respRef = ref
				if c := parseCode(code); c > 0 {
					statusCode = c
				}
				break
			}
		}
	}

	if respRef == nil || respRef.Value == nil || respRef.Value.Content == nil {
		return statusCode, map[string]any{}, nil
	}

	mediaType := respRef.Value.Content.Get("application/json")
	if mediaType == nil || mediaType.Schema == nil || mediaType.Schema.Value == nil {
		// Fallback to any content type if JSON not present
		for _, mt := range respRef.Value.Content {
			if mt != nil && mt.Schema != nil && mt.Schema.Value != nil {
				mediaType = mt
				break
			}
		}
	}

	if mediaType == nil || mediaType.Schema == nil || mediaType.Schema.Value == nil {
		return statusCode, map[string]any{}, nil
	}

	seed := GenerateSeed(httpMethod, path, operationID)
	ctx := NewGenerationContext(seed, maxDepth)

	payload, err := w.generateSchema(ctx, mediaType.Schema.Value)
	if err != nil {
		return statusCode, nil, err
	}

	return statusCode, payload, nil
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
			val, err := w.generateSchema(ctx, propRef.Value)
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
			count = int(schema.MinItems)
		}
		if schema.MaxItems != nil && *schema.MaxItems > 0 {
			max := int(*schema.MaxItems)
			if max < count {
				max = count
			}
			count = max
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

func parseCode(code string) int {
	if code == "200" {
		return 200
	}
	if code == "201" {
		return 201
	}
	if code == "202" {
		return 202
	}
	if code == "204" {
		return 204
	}
	return 200
}
