package memory

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
)

type MilvusStore struct {
	client     client.Client
	collection string
	dimension  int
	metricType entity.MetricType
}

type MilvusStoreConfig struct {
	Collection string
	Dimension  int
	MetricType string
}

func NewMilvusStore(cli client.Client, cfg MilvusStoreConfig) *MilvusStore {
	if cli == nil || strings.TrimSpace(cfg.Collection) == "" {
		return nil
	}
	return &MilvusStore{
		client:     cli,
		collection: cfg.Collection,
		dimension:  cfg.Dimension,
		metricType: parseMetricType(cfg.MetricType),
	}
}

func parseMetricType(metric string) entity.MetricType {
	switch strings.ToUpper(strings.TrimSpace(metric)) {
	case "L2":
		return entity.L2
	case "IP":
		return entity.IP
	case "COSINE":
		return entity.COSINE
	default:
		return entity.COSINE
	}
}

func (s *MilvusStore) ensureCollection(ctx context.Context, dimension int) error {
	if s == nil || s.client == nil {
		return fmt.Errorf("milvus client not initialized")
	}
	if s.dimension == 0 {
		s.dimension = dimension
	}
	if s.dimension == 0 {
		return fmt.Errorf("milvus dimension not set")
	}
	has, err := s.client.HasCollection(ctx, s.collection)
	if err != nil {
		return err
	}
	if has {
		if err := s.verifyCollection(ctx); err != nil {
			return err
		}
		return s.client.LoadCollection(ctx, s.collection, false)
	}

	schema := entity.NewSchema().
		WithName(s.collection).
		WithDescription("memories").
		WithAutoID(false).
		WithDynamicFieldEnabled(false)

	pkField := entity.NewField().
		WithName("id").
		WithDataType(entity.FieldTypeVarChar).
		WithIsPrimaryKey(true).
		WithTypeParams(entity.TypeParamMaxLength, "64")

	personaField := entity.NewField().
		WithName("persona_id").
		WithDataType(entity.FieldTypeVarChar).
		WithTypeParams(entity.TypeParamMaxLength, "64")

	userField := entity.NewField().
		WithName("user_id").
		WithDataType(entity.FieldTypeInt64)

	vectorField := entity.NewField().
		WithName("embedding").
		WithDataType(entity.FieldTypeFloatVector).
		WithTypeParams(entity.TypeParamDim, strconv.Itoa(s.dimension))

	schema.WithField(pkField).
		WithField(personaField).
		WithField(userField).
		WithField(vectorField)

	if err := s.client.CreateCollection(ctx, schema, 1); err != nil {
		return err
	}
	index, err := entity.NewIndexFlat(s.metricType)
	if err != nil {
		return err
	}
	if err := s.client.CreateIndex(ctx, s.collection, "embedding", index, false); err != nil {
		return err
	}
	return s.client.LoadCollection(ctx, s.collection, false)
}

func (s *MilvusStore) verifyCollection(ctx context.Context) error {
	collection, err := s.client.DescribeCollection(ctx, s.collection)
	if err != nil {
		return err
	}
	for _, field := range collection.Schema.Fields {
		if field.Name != "embedding" {
			continue
		}
		dimStr := field.TypeParams[entity.TypeParamDim]
		if dimStr == "" {
			return fmt.Errorf("milvus embedding dim missing")
		}
		dim, err := strconv.Atoi(dimStr)
		if err != nil {
			return err
		}
		if s.dimension == 0 {
			s.dimension = dim
		} else if s.dimension != dim {
			return fmt.Errorf("milvus embedding dim mismatch: %d vs %d", s.dimension, dim)
		}
		return nil
	}
	return fmt.Errorf("milvus embedding field not found")
}

func (s *MilvusStore) UpsertMemory(ctx context.Context, memoryID, personaID string, userID int64, embedding []float32) error {
	if s == nil || s.client == nil {
		return fmt.Errorf("milvus client not initialized")
	}
	if memoryID == "" || personaID == "" || userID == 0 {
		return fmt.Errorf("invalid memory identity")
	}
	if len(embedding) == 0 {
		return fmt.Errorf("empty embedding")
	}
	if err := s.ensureCollection(ctx, len(embedding)); err != nil {
		return err
	}
	if s.dimension != len(embedding) {
		return fmt.Errorf("embedding dim mismatch: %d vs %d", s.dimension, len(embedding))
	}

	expr := fmt.Sprintf("id == \"%s\"", escapeExprString(memoryID))
	_ = s.client.Delete(ctx, s.collection, "", expr)

	columns := []entity.Column{
		entity.NewColumnVarChar("id", []string{memoryID}),
		entity.NewColumnVarChar("persona_id", []string{personaID}),
		entity.NewColumnInt64("user_id", []int64{userID}),
		entity.NewColumnFloatVector("embedding", s.dimension, [][]float32{embedding}),
	}
	_, err := s.client.Insert(ctx, s.collection, "", columns...)
	return err
}

func (s *MilvusStore) DeleteMemory(ctx context.Context, memoryID string) error {
	if s == nil || s.client == nil || memoryID == "" {
		return nil
	}
	expr := fmt.Sprintf("id == \"%s\"", escapeExprString(memoryID))
	return s.client.Delete(ctx, s.collection, "", expr)
}

func (s *MilvusStore) SearchMemories(ctx context.Context, personaID string, userID int64, embedding []float32, topK int) ([]string, error) {
	if s == nil || s.client == nil {
		return nil, fmt.Errorf("milvus client not initialized")
	}
	if topK <= 0 || len(embedding) == 0 {
		return []string{}, nil
	}
	if err := s.ensureCollection(ctx, len(embedding)); err != nil {
		return nil, err
	}
	if s.dimension != len(embedding) {
		return nil, fmt.Errorf("embedding dim mismatch: %d vs %d", s.dimension, len(embedding))
	}
	sp, err := entity.NewIndexFlatSearchParam()
	if err != nil {
		return nil, err
	}
	expr := fmt.Sprintf("persona_id == \"%s\" && user_id == %d", escapeExprString(personaID), userID)
	vectors := []entity.Vector{entity.FloatVector(embedding)}
	results, err := s.client.Search(ctx, s.collection, nil, expr, []string{}, vectors, "embedding", s.metricType, topK, sp)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 || results[0].ResultCount == 0 {
		return []string{}, nil
	}
	idsColumn, ok := results[0].IDs.(*entity.ColumnVarChar)
	if !ok {
		return nil, fmt.Errorf("unexpected id column type")
	}
	return idsColumn.Data(), nil
}

func escapeExprString(input string) string {
	escaped := strings.ReplaceAll(input, "\\", "\\\\")
	return strings.ReplaceAll(escaped, "\"", "\\\"")
}
