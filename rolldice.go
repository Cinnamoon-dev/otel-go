package main

import (
	"io"
	"log"
	"math/rand"
	"net/http"
	"strconv"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

const name = "go.opentelemetry.io/contrib/examples/dice"

var (
	logger = otelslog.NewLogger(name)
)

type Telemetry struct {
	Tracer  trace.Tracer
	Meter   metric.Meter
	RollCnt metric.Int64Counter
}

func NewTelemetry() (*Telemetry, error) {
	meter := otel.Meter(name)
	rollCnt, err := meter.Int64Counter(
		"dice.rolls",
		metric.WithDescription("The number of rolls by roll value"),
		metric.WithUnit("{roll}"),
	)
	if err != nil {
		return nil, err
	}

	return &Telemetry{
		Tracer:  otel.Tracer(name),
		Meter:   meter,
		RollCnt: rollCnt,
	}, nil
}

func rolldice(w http.ResponseWriter, r *http.Request) {
	telemetry, err := NewTelemetry()
	if err != nil {
		panic(err)
	}

	ctx, span := telemetry.Tracer.Start(r.Context(), "roll")
	defer span.End()

	roll := 1 + rand.Intn(6)

	var msg string
	if player := r.PathValue("player"); player != "" {
		msg = player + " is rolling the dice"
		rollPlayerAttr := attribute.String("roll.player", player)
		span.SetAttributes(rollPlayerAttr)
	} else {
		msg = "Anonymous player is rolling the dice"
		rollPlayerAttr := attribute.String("roll.player", "Anonymous")
		span.SetAttributes(rollPlayerAttr)
	}
	logger.InfoContext(ctx, msg, "result", roll)

	rollValueAttr := attribute.Int("roll.value", roll)
	span.SetAttributes(rollValueAttr)
	telemetry.RollCnt.Add(ctx, 1, metric.WithAttributes(rollValueAttr))

	log.Printf("%s, result: %d", msg, roll)

	resp := strconv.Itoa(roll) + "\n"
	if _, err := io.WriteString(w, resp); err != nil {
		log.Printf("Write failed: %v", err)
	}
}
