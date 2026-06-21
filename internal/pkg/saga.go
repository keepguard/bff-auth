package pkg

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// Step representa um passo de uma saga
type Step struct {
	Name    string
	Do      func(ctx context.Context) error
	Undo    func(ctx context.Context) error
	Timeout time.Duration
}

// Saga representa uma saga orquestrada
type Saga struct {
	name   string
	steps  []Step
	logger *zap.Logger
}

// NewSaga cria uma nova saga
func NewSaga(name string) *Saga {
	return &Saga{
		name:   name,
		steps:  make([]Step, 0),
		logger: zap.NewNop(), // Logger padrão que não faz nada
	}
}

// WithLogger adiciona um logger à saga
func (s *Saga) WithLogger(logger *zap.Logger) *Saga {
	s.logger = logger
	return s
}

// AddStep adiciona um passo à saga
func (s *Saga) AddStep(name string, do, undo func(ctx context.Context) error, timeout time.Duration) *Saga {
	s.steps = append(s.steps, Step{
		Name:    name,
		Do:      do,
		Undo:    undo,
		Timeout: timeout,
	})
	return s
}

// Execute executa a saga
func (s *Saga) Execute(ctx context.Context) error {
	executedSteps := make([]Step, 0)

	// Executa todos os passos
	for i, step := range s.steps {
		stepCtx := ctx
		if step.Timeout > 0 {
			var cancel context.CancelFunc
			stepCtx, cancel = context.WithTimeout(ctx, step.Timeout)
			defer cancel()
		}

		if err := step.Do(stepCtx); err != nil {
			// Em caso de erro, executa compensação em ordem inversa
			s.compensate(ctx, executedSteps)
			return fmt.Errorf("erro no passo %d (%s): %w", i+1, step.Name, err)
		}

		executedSteps = append(executedSteps, step)
	}

	return nil
}

// compensate executa as ações de compensação em ordem inversa
func (s *Saga) compensate(ctx context.Context, executedSteps []Step) {
	for i := len(executedSteps) - 1; i >= 0; i-- {
		step := executedSteps[i]
		if step.Undo != nil {
			stepCtx := ctx
			if step.Timeout > 0 {
				var cancel context.CancelFunc
				stepCtx, cancel = context.WithTimeout(ctx, step.Timeout)
				defer cancel()
			}

			if err := step.Undo(stepCtx); err != nil {
				// Log do erro de compensação, mas não falha a saga
				s.logger.Error("Erro na compensação do passo da saga",
					zap.String("saga", s.name),
					zap.String("step", step.Name),
					zap.Error(err),
				)
			}
		}
	}
}
