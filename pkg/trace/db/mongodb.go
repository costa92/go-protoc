package db

import (
	"context"
	"fmt"
	"time"

	"github.com/costa92/go-protoc/v2/pkg/trace"
	"go.mongodb.org/mongo-driver/event"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.opentelemetry.io/contrib/instrumentation/go.mongodb.org/mongo-driver/mongo/otelmongo"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	ottrace "go.opentelemetry.io/otel/trace"
)

const (
	MongoDBTracerName = "mongodb"
)

// NewMongoMonitor creates a new MongoDB trace monitor using the official OpenTelemetry instrumentation.
func NewMongoMonitor(opts ...otelmongo.Option) *event.CommandMonitor {
	return otelmongo.NewMonitor(opts...)
}

// NewMongoMonitorWithOptions creates a MongoDB trace monitor with custom options.
func NewMongoMonitorWithOptions() *event.CommandMonitor {
	return otelmongo.NewMonitor()
}

// NewTracedMongoClient creates a new MongoDB client with OpenTelemetry tracing enabled.
func NewTracedMongoClient(ctx context.Context, uri string, mongoOpts ...otelmongo.Option) (*mongo.Client, error) {
	// Set up the monitor
	monitor := NewMongoMonitor(mongoOpts...)

	// Configure client options with the monitor
	clientOpts := options.Client().ApplyURI(uri).SetMonitor(monitor)

	// Connect to MongoDB
	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Test the connection
	if err := client.Ping(ctx, nil); err != nil {
		client.Disconnect(ctx)
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	return client, nil
}

// TracedMongoCollection wraps a MongoDB collection with tracing utilities.
type TracedMongoCollection struct {
	*mongo.Collection
	database   string
	collection string
}

// NewTracedMongoCollection creates a new traced MongoDB collection.
func NewTracedMongoCollection(client *mongo.Client, database, collection string) *TracedMongoCollection {
	return &TracedMongoCollection{
		Collection: client.Database(database).Collection(collection),
		database:   database,
		collection: collection,
	}
}

// WithTracing wraps a MongoDB operation with additional tracing context.
func (c *TracedMongoCollection) WithTracing(ctx context.Context, operation string) (context.Context, ottrace.Span) {
	spanName := fmt.Sprintf("mongodb.%s", operation)
	tracer := trace.GetTracer(MongoDBTracerName)
	ctx, span := tracer.Start(ctx, spanName, trace.WithSpanKind(ottrace.SpanKindClient))

	span.SetAttributes(
		attribute.String("db.system", "mongodb"),
		attribute.String("db.operation", operation),
		attribute.String("db.name", c.database),
		attribute.String("db.mongodb.collection", c.collection),
	)

	return ctx, span
}

// TraceableTransaction provides transaction support with tracing.
type TraceableTransaction struct {
	session mongo.Session
	ctx     context.Context
	span    ottrace.Span
}

// StartTracedTransaction starts a MongoDB transaction with tracing.
func StartTracedTransaction(ctx context.Context, client *mongo.Client) (*TraceableTransaction, error) {
	tracer := trace.GetTracer(MongoDBTracerName)
	ctx, span := tracer.Start(ctx, "mongodb.transaction", trace.WithSpanKind(ottrace.SpanKindClient))
	span.SetAttributes(
		attribute.String("db.system", "mongodb"),
		attribute.String("db.operation", "transaction"),
	)

	session, err := client.StartSession()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		span.End()
		return nil, err
	}

	return &TraceableTransaction{
		session: session,
		ctx:     ctx,
		span:    span,
	}, nil
}

// Commit commits the traced MongoDB transaction.
func (t *TraceableTransaction) Commit() error {
	defer t.span.End()
	defer t.session.EndSession(t.ctx)

	if err := t.session.CommitTransaction(t.ctx); err != nil {
		t.span.RecordError(err)
		t.span.SetStatus(codes.Error, err.Error())
		return err
	}

	t.span.SetStatus(codes.Ok, "committed")
	return nil
}

// Abort aborts the traced MongoDB transaction.
func (t *TraceableTransaction) Abort() error {
	defer t.span.End()
	defer t.session.EndSession(t.ctx)

	if err := t.session.AbortTransaction(t.ctx); err != nil {
		t.span.RecordError(err)
		t.span.SetStatus(codes.Error, err.Error())
		return err
	}

	t.span.SetStatus(codes.Ok, "aborted")
	return nil
}

// Session returns the MongoDB session for use in operations.
func (t *TraceableTransaction) Session() mongo.Session {
	return t.session
}

// Context returns the traced context.
func (t *TraceableTransaction) Context() context.Context {
	return t.ctx
}

// MongoWithContext creates a MongoDB operation with context for tracing.
func MongoWithContext(ctx context.Context, operation string, fn func(context.Context) error) error {
	tracer := trace.GetTracer(MongoDBTracerName)
	ctx, span := tracer.Start(ctx, fmt.Sprintf("mongodb.%s", operation), trace.WithSpanKind(ottrace.SpanKindClient))
	defer span.End()

	span.SetAttributes(
		attribute.String("db.system", "mongodb"),
		attribute.String("db.operation", operation),
	)

	start := time.Now()
	err := fn(ctx)
	duration := time.Since(start)

	span.SetAttributes(
		attribute.Int64("mongodb.duration", duration.Microseconds()),
	)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	} else {
		span.SetStatus(codes.Ok, "")
	}

	return err
}
