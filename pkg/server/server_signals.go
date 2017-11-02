package server

import (
	"os/signal"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
)

func (s *Server) configureSignals() {
	signal.Notify(s.signals, syscall.SIGINT, syscall.SIGTERM)
}

func (s *Server) listenSignals() {
	for {
		sig := <-s.signals
		switch sig {
		default:
			log.WithField("signal", sig).Info("I have to go...")
			reqAcceptGraceTimeOut := time.Duration(s.globalConfiguration.GraceTimeOut)
			if reqAcceptGraceTimeOut > 0 {
				log.WithField("timeout", s.globalConfiguration.GraceTimeOut).Info("Waiting %s for incoming requests to cease")
				time.Sleep(reqAcceptGraceTimeOut)
			}
			log.Info("Stopping server gracefully")
			s.Stop()
		}
	}
}
