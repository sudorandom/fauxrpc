package generator

import (
	"math/rand"
	"sort"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

type PolyResolver struct{}

func NewPolyResolver() *PolyResolver {
	return &PolyResolver{}
}

// ResolveAllOf evaluates all subschemas in an allOf array and merges their properties.
func (p *PolyResolver) ResolveAllOf(walker *Walker, ctx *GenerationContext, allOf []*openapi3.SchemaRef) (any, error) {
	merged := make(map[string]any)

	for _, ref := range allOf {
		if ref == nil || ref.Value == nil {
			continue
		}
		subVal, err := walker.generateSchema(ctx, ref.Value)
		if err != nil {
			return nil, err
		}
		if subMap, ok := subVal.(map[string]any); ok {
			for k, v := range subMap {
				merged[k] = v // Conflict handling: last subschema overrides earlier properties
			}
		}
	}

	return merged, nil
}

// ResolveOneOfOrAnyOf evaluates oneOf or anyOf using discriminator or deterministic PRNG index.
func (p *PolyResolver) ResolveOneOfOrAnyOf(walker *Walker, ctx *GenerationContext, list []*openapi3.SchemaRef, discriminator *openapi3.Discriminator) (any, error) {
	if len(list) == 0 {
		return nil, nil
	}

	// 1. Discriminator check
	if discriminator != nil && discriminator.PropertyName != "" {
		// If discriminator points to an explicit mapping or schema name
		if len(discriminator.Mapping) > 0 {
			discriminatorValues := make([]string, 0, len(discriminator.Mapping))
			for discriminatorValue := range discriminator.Mapping {
				discriminatorValues = append(discriminatorValues, discriminatorValue)
			}
			sort.Strings(discriminatorValues)
			for _, discVal := range discriminatorValues {
				mappingRef := discriminator.Mapping[discVal]
				encodedRef, err := mappingRef.MarshalText()
				if err != nil {
					continue
				}
				targetRef := string(encodedRef)
				for _, ref := range list {
					if discriminatorRefsMatch(ref.Ref, targetRef) {
						selectedRef := ref
						res, err := p.evalBranch(walker, ctx, selectedRef)
						if err != nil {
							return nil, err
						}
						if resMap, ok := res.(map[string]any); ok {
							resMap[discriminator.PropertyName] = discVal
							return resMap, nil
						}
						return res, nil
					}
				}
			}
		}
	}

	// 2. Deterministic PRNG selection
	r := rand.New(rand.NewSource(ctx.seed))
	idx := r.Intn(len(list))
	selectedRef := list[idx]

	return p.evalBranch(walker, ctx, selectedRef)
}

func discriminatorRefsMatch(branchRef, targetRef string) bool {
	if branchRef == targetRef {
		return true
	}
	if strings.Contains(targetRef, "/") {
		return false
	}
	return strings.TrimPrefix(branchRef, "#/components/schemas/") == targetRef
}

func (p *PolyResolver) evalBranch(walker *Walker, ctx *GenerationContext, ref *openapi3.SchemaRef) (any, error) {
	if ref == nil || ref.Value == nil {
		return nil, nil
	}

	// Evaluate a copy with additional properties explicitly disabled. Keeping the
	// parsed schema immutable avoids races when requests generate concurrently.
	branch := *ref.Value
	allowAdditionalProperties := false
	branch.AdditionalProperties.Has = &allowAdditionalProperties

	return walker.generateSchema(ctx, &branch)
}
