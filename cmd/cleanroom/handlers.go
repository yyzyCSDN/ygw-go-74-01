package main

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"cleanroomorcontrol/internal/model"
	"cleanroomorcontrol/internal/service"
)

func registerHandlers(mux *http.ServeMux, svc *service.Service, webDir string) {
	mux.HandleFunc("/healthz", healthHandler)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		server := &Server{service: svc, webDir: webDir}
		server.serveConsole(w, r)
	})
	mux.HandleFunc("/api/v1/status", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, svc.Status())
	})
	mux.HandleFunc("/api/v1/pressure/sample", func(w http.ResponseWriter, r *http.Request) {
		var reading model.PressureReading
		if !readJSON(w, r, &reading) {
			return
		}
		reading = reading.Normalized()
		status := svc.HandlePressureSample(reading)
		writeJSON(w, http.StatusOK, map[string]any{"room": reading.Room, "status": status.String(), "pa": reading.Pa})
	})
	mux.HandleFunc("/api/v1/particle/sample", func(w http.ResponseWriter, r *http.Request) {
		var sample model.ParticleSample
		if !readJSON(w, r, &sample) {
			return
		}
		sample = sample.Normalized()
		if err := svc.HandleParticleSample(sample); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"point": sample.Point, "count": sample.Count})
	})
	mux.HandleFunc("/api/v1/particle/cycle", func(w http.ResponseWriter, r *http.Request) {
		var req particleCycleRequest
		if !readJSON(w, r, &req) {
			return
		}
		result := svc.RunParticleCycle(req.Samples)
		writeJSON(w, http.StatusOK, map[string]any{"room": result.Room, "samples": result.Samples, "errors": result.Errors})
	})
	mux.HandleFunc("/api/v1/disinfection/start", func(w http.ResponseWriter, r *http.Request) {
		var req roomRequest
		if !readJSON(w, r, &req) {
			return
		}
		if err := svc.HandleDisinfectionStart(req.Room); err != nil {
			writeJSON(w, http.StatusConflict, map[string]any{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"room": req.Room, "phase": svc.Disinfection.Phase(req.Room).String()})
	})
	mux.HandleFunc("/api/v1/disinfection/schedule", func(w http.ResponseWriter, r *http.Request) {
		var req scheduleRequest
		if !readJSON(w, r, &req) {
			return
		}
		svc.ScheduleDisinfection(req.Room, req.At)
		writeJSON(w, http.StatusOK, map[string]any{"room": req.Room, "at": req.At})
	})
	mux.HandleFunc("/api/v1/disinfection/complete", func(w http.ResponseWriter, r *http.Request) {
		var req roomRequest
		if !readJSON(w, r, &req) {
			return
		}
		if err := svc.HandleDisinfectionComplete(req.Room); err != nil {
			writeJSON(w, http.StatusConflict, map[string]any{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"room": req.Room, "phase": svc.Disinfection.Phase(req.Room).String()})
	})
	mux.HandleFunc("/api/v1/ventilation", func(w http.ResponseWriter, r *http.Request) {
		var req roomRequest
		if !readJSON(w, r, &req) {
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()
		if err := svc.HandleVentilation(ctx, req.Room); err != nil {
			writeJSON(w, http.StatusConflict, map[string]any{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"room": req.Room, "phase": svc.Disinfection.Phase(req.Room).String()})
	})
	mux.HandleFunc("/api/v1/fan/switch", func(w http.ResponseWriter, r *http.Request) {
		var req fanSwitchRequest
		if !readJSON(w, r, &req) {
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()
		if err := svc.HandleFanSwitch(ctx, req.Room, req.Role, req.Target); err != nil {
			writeJSON(w, http.StatusConflict, map[string]any{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"room": req.Room, "target": req.Target})
	})
	mux.HandleFunc("/api/v1/fan/fail", func(w http.ResponseWriter, r *http.Request) {
		var req fanFailRequest
		if !readJSON(w, r, &req) {
			return
		}
		svc.Driver.FailSupply(req.Fan)
		writeJSON(w, http.StatusOK, map[string]any{"fan": req.Fan, "failed": true})
	})
	mux.HandleFunc("/api/v1/airflow/purge", func(w http.ResponseWriter, r *http.Request) {
		var req roomRequest
		if !readJSON(w, r, &req) {
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()
		if err := svc.PurgeRoom(ctx, req.Room); err != nil {
			writeJSON(w, http.StatusConflict, map[string]any{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"room": req.Room, "mode": svc.Airflow.Mode(req.Room).String()})
	})
	mux.HandleFunc("/api/v1/airflow/standby", func(w http.ResponseWriter, r *http.Request) {
		var req roomRequest
		if !readJSON(w, r, &req) {
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()
		if err := svc.StandbyRoom(ctx, req.Room); err != nil {
			writeJSON(w, http.StatusConflict, map[string]any{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"room": req.Room, "mode": svc.Airflow.Mode(req.Room).String()})
	})
	mux.HandleFunc("/api/v1/door/acquire", func(w http.ResponseWriter, r *http.Request) {
		var req roomRequest
		if !readJSON(w, r, &req) {
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()
		if err := svc.DoorAcquire(ctx, req.Room); err != nil {
			writeJSON(w, http.StatusConflict, map[string]any{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"room": req.Room, "door": svc.Door.State(req.Room).String()})
	})
	mux.HandleFunc("/api/v1/particle/point/register", func(w http.ResponseWriter, r *http.Request) {
		var req pointRequest
		if !readJSON(w, r, &req) {
			return
		}
		svc.RegisterParticlePoint(req.Point, req.Room, req.Limit)
		writeJSON(w, http.StatusOK, map[string]any{"point": req.Point, "room": req.Room, "limit": req.Limit})
	})
	mux.HandleFunc("/api/v1/alarms/active", func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		writeJSON(w, http.StatusOK, map[string]any{"key": key, "active": svc.AlarmActive(key)})
	})
	mux.HandleFunc("/api/v1/env/sample", func(w http.ResponseWriter, r *http.Request) {
		var reading model.EnvReading
		if !readJSON(w, r, &reading) {
			return
		}
		reading = reading.Normalized()
		violations := svc.HandleEnvSample(reading)
		writeJSON(w, http.StatusOK, map[string]any{"room": reading.Room, "violations": violations})
	})
	mux.HandleFunc("/api/v1/alarms", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"alarms": svc.AlarmList()})
	})
	mux.HandleFunc("/api/v1/doors", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"doors": svc.DoorStates()})
	})
	mux.HandleFunc("/api/v1/persist", func(w http.ResponseWriter, r *http.Request) {
		if err := svc.Persist(); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"persisted": true})
	})
	mux.HandleFunc("/api/v1/settle", func(w http.ResponseWriter, r *http.Request) {
		var req roomRequest
		if !readJSON(w, r, &req) {
			return
		}
		svc.Signals.MarkSettled(req.Room)
		writeJSON(w, http.StatusOK, map[string]any{"room": req.Room, "settled": true})
	})
}

type roomRequest struct {
	Room model.RoomID
}

type fanSwitchRequest struct {
	Room   model.RoomID
	Role   model.FanRole
	Target model.FanID
}

type particleCycleRequest struct {
	Samples []model.ParticleSample
}

type fanFailRequest struct {
	Fan model.FanID
}

type scheduleRequest struct {
	Room model.RoomID
	At   time.Time
}

type pointRequest struct {
	Point model.PointID
	Room  model.RoomID
	Limit int
}

func readJSON(w http.ResponseWriter, r *http.Request, target any) bool {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "method not allowed"})
		return false
	}
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(target); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return false
	}
	return true
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
