package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"cloud.google.com/go/datastore"
	texporter "github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace"
	"github.com/google/uuid"
	octrace "go.opencensus.io/trace"
	"go.opentelemetry.io/contrib/detectors/gcp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/bridge/opencensus"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

type Entity struct {
	Value     string
	CreatedAt time.Time
}

func main() {
	// Start Tracing
	tp, err := InitTracer()
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Printf("Error shutting down tracer provider: %v", err)
		}
	}()

	fmt.Println("Hello, Kubernetes Novice!")

	tracer := otel.GetTracerProvider().Tracer("go.opentelemetry.io/otel/bridge/opencensus")
	octrace.DefaultTracer = opencensus.NewTracer(tracer)

	ctx := context.Background()
	ctx, span := tracer.Start(ctx, "start")
	defer span.End()

	projectID := "nttd-platformtec"
	fmt.Println("projectID: ", projectID)

	// clinet 接続から datastore 操作, close までを無限ループで実行
	client, err := datastore.NewClient(
		ctx,
		projectID,
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	uuid := uuid.New().String()
	key := datastore.NameKey("entity_k8snovice", uuid, nil)
	e := new(Entity)
	randomEmoji := getEmoji()
	e.Value = "Hi! Kubernetes Novice " + randomEmoji
	loc, _ := time.LoadLocation("Asia/Tokyo")
	e.CreatedAt = time.Now().In(loc)

	if _, err := client.Put(ctx, key, e); err != nil {
		log.Fatalf("Failed to save entity: %v", err)
	}

	e = new(Entity)
	if err = client.Get(ctx, key, e); err != nil {
		log.Fatalf("Failed to get entity: %v", err)
	}

	_, span = tracer.Start(ctx, "finish")
	defer span.End()
	fmt.Printf("Fetched entity: %v", e)

	client.Close()
}

func getEmoji() string {
	emojis := []string{"😀", "😃", "😄", "😁", "😆", "😅", "😂", "🤣", "😊", "😇"}
	rand.Seed(time.Now().UnixNano())
	randomEmoji := emojis[rand.Intn(len(emojis))]

	return randomEmoji
}

func InitTracer() (*sdktrace.TracerProvider, error) {
	ctx := context.Background()

	projectID := "nttd-platformtec"
	exporter, err := texporter.New(texporter.WithProjectID(projectID))
	if err != nil {
		log.Fatalf("texporter.New: %v", err)
	}

	res, err := resource.New(ctx,
		// Use the GCP resource detector to detect information about the GCP platform
		resource.WithDetectors(gcp.NewDetector()),
		// Keep the default detectors
		resource.WithTelemetrySDK(),
		// Add your own custom attributes to identify your application
		resource.WithAttributes(
			semconv.ServiceNameKey.String("my-application"),
		),
	)
	if err != nil {
		log.Fatalf("resource.New: %v", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)
	defer tp.Shutdown(ctx) // flushes any pending spans, and closes connections.
	otel.SetTracerProvider(tp)

	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return tp, nil
}
