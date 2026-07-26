package generator

import (
	"math/rand"

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
		// If discriminator points to a explicit mapping or schema name
		if len(discriminator.Mapping) > 0 {
			// Find mapped schema
			for discVal, targetRef := range discriminator.Mapping {
				for _, ref := range list {
					if ref.Ref == targetRef.Ref || ref.Ref == "#/components/schemas/"+targetRef.Ref {
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

func (p *PolyResolver) evalBranch(walker *Walker, ctx *GenerationContext, ref *openapi3.SchemaRef) (any, error) {
	if ref == nil || ref.Value == nil {
		return nil, nil
	}

	// Temporarily set AdditionalPropertiesAllowed = false on branch schema
	origAddProps := ref.Value.AdditionalProperties.Has
	ref.Value.AdditionalProperties.Has = nil
	defer func() {
		ref.Value.AdditionalProperties.Has = origAddProps
	}()

	return walker.generateSchema(ctx, ref.Value)
}
