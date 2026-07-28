package generator

import (
	"math/rand"
	"testing"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScalarSynthesizerGeneratesDateTime(t *testing.T) {
	schema := openapi3.NewDateTimeSchema()
	synthesizer := NewScalarSynthesizer()

	value := synthesizer.Synthesize(schema, rand.New(rand.NewSource(42)))
	dateTime, ok := value.(string)
	require.True(t, ok)
	parsed, err := time.Parse(time.RFC3339, dateTime)
	require.NoError(t, err)
	assert.False(t, parsed.Before(syntheticTimeStart))
	assert.True(t, parsed.Before(syntheticTimeEnd))
	assert.Equal(t, dateTime, synthesizer.Synthesize(schema, rand.New(rand.NewSource(42))))
	assert.NotEqual(t, dateTime, synthesizer.Synthesize(schema, rand.New(rand.NewSource(43))))
}

func TestScalarSynthesizerGeneratesSeededUUID(t *testing.T) {
	schema := openapi3.NewUUIDSchema()
	synthesizer := NewScalarSynthesizer()

	first := synthesizer.Synthesize(schema, rand.New(rand.NewSource(42)))
	second := synthesizer.Synthesize(schema, rand.New(rand.NewSource(42)))
	differentSeed := synthesizer.Synthesize(schema, rand.New(rand.NewSource(43)))

	assert.Equal(t, first, second)
	assert.NotEqual(t, first, differentSeed)
	_, err := uuid.Parse(first.(string))
	require.NoError(t, err)
}

func TestScalarSynthesizerGeneratesSeededString(t *testing.T) {
	maxLength := uint64(20)
	schema := openapi3.NewStringSchema()
	schema.MinLength = 12
	schema.MaxLength = &maxLength
	synthesizer := NewScalarSynthesizer()

	first := synthesizer.Synthesize(schema, rand.New(rand.NewSource(42)))
	second := synthesizer.Synthesize(schema, rand.New(rand.NewSource(42)))
	differentSeed := synthesizer.Synthesize(schema, rand.New(rand.NewSource(43)))

	assert.Equal(t, first, second)
	assert.NotEqual(t, first, differentSeed)
	value := first.(string)
	assert.GreaterOrEqual(t, len(value), 12)
	assert.LessOrEqual(t, len(value), 20)
}

func TestScalarSynthesizerGeneratesDate(t *testing.T) {
	schema := openapi3.NewStringSchema()
	schema.Format = "date"
	synthesizer := NewScalarSynthesizer()

	value := synthesizer.Synthesize(schema, rand.New(rand.NewSource(42)))
	date, ok := value.(string)
	require.True(t, ok)
	_, err := time.Parse("2006-01-02", date)
	require.NoError(t, err)
	assert.Equal(t, date, synthesizer.Synthesize(schema, rand.New(rand.NewSource(42))))
}
