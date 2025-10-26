package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/kunalkushwaha/agenticgokit/core"
)

func main() {
	cfg, err := core.LoadConfigFromWorkingDir()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	provider, err := cfg.InitializeProvider()
	log.Printf("Provider %v", &provider)

	if err != nil {
		log.Fatalf("Failed to create LLM provider: %v", err)
	}

	// 🤖 Create three specialized agents
	agents := map[string]core.AgentHandler{
		"processor": &ProcessorAgent{llm: provider},
		"enhancer":  &EnhancerAgent{llm: provider},
		"formatter": &FormatterAgent{llm: provider},
	}

	// 🚀 Prefer config-driven runner (supports route/collab/seq/loop/mixed)
	runner, err := core.NewRunnerFromConfig("agentflow.toml")
	if err != nil {
		log.Fatalf("runner: %v", err)
	}

	// 💬 Process a message - watch the magic happen!
	fmt.Println("🤖 Starting multi-agent collaboration...")

	// Start the runner
	ctx := context.Background()
	runner.Start(ctx)
	defer runner.Stop()

	// Create an event for processing
	event := core.NewEvent("processor", core.EventData{
		"input": "Explain quantum computing in simple terms",
	}, map[string]string{
		"route": "processor",
	})

	// Emit the event to the runner
	if err := runner.Emit(event); err != nil {
		log.Fatalf("Failed to emit event: %v", err)
	}

	// Wait for processing to complete
	time.Sleep(5 * time.Second)

	fmt.Println("\n✅ Multi-Agent Processing Complete!")
	fmt.Println("=" + strings.Repeat("=", 50))
	fmt.Printf("📊 Execution Stats:\n")
	fmt.Printf("   • Agents involved: %d\n", len(agents))
	fmt.Printf("   • Event ID: %s\n", event.GetID())
}

// ProcessorAgent handles initial processing
type ProcessorAgent struct {
	llm core.ModelProvider
}

// EnhancerAgent enhances the processed information
type EnhancerAgent struct {
	llm core.ModelProvider
}

// FormatterAgent formats the final response
type FormatterAgent struct {
	llm core.ModelProvider
}

func (a *ProcessorAgent) Run(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
	// Get user input from event data
	input, ok := event.GetData()["input"].(string)
	if !ok {
		return core.AgentResult{}, fmt.Errorf("no input provided")
	}

	// Process with LLM
	prompt := core.Prompt{
		System: "You are a processor agent. Extract and organize key information from user requests.",
		User:   fmt.Sprintf("Process this request and extract key information: %s", input),
	}

	response, err := a.llm.Call(ctx, prompt)
	if err != nil {
		return core.AgentResult{}, err
	}

	// Update state with processed result
	outputState := core.NewState()
	outputState.Set("processed", response.Content)
	outputState.Set("message", response.Content)

	// Route to enhancer
	outputState.SetMeta(core.RouteMetadataKey, "enhancer")

	return core.AgentResult{OutputState: outputState}, nil
}

func (a *EnhancerAgent) Run(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
	// Get processed result from state
	var processed interface{}
	if processedData, exists := state.Get("processed"); exists {
		processed = processedData
	} else if msg, exists := state.Get("message"); exists {
		processed = msg
	} else {
		return core.AgentResult{}, fmt.Errorf("no processed data found")
	}

	// Enhance with LLM
	prompt := core.Prompt{
		System: "You are an enhancer agent. Add insights, context, and additional valuable information.",
		User:   fmt.Sprintf("Enhance this response with additional insights: %v", processed),
	}

	response, err := a.llm.Call(ctx, prompt)
	if err != nil {
		return core.AgentResult{}, err
	}

	// Update state with enhanced result
	outputState := core.NewState()
	outputState.Set("enhanced", response.Content)
	outputState.Set("message", response.Content)

	// Route to formatter
	outputState.SetMeta(core.RouteMetadataKey, "formatter")

	return core.AgentResult{OutputState: outputState}, nil
}

func (a *FormatterAgent) Run(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
	// Get enhanced result from state
	var enhanced interface{}
	if enhancedData, exists := state.Get("enhanced"); exists {
		enhanced = enhancedData
	} else if msg, exists := state.Get("message"); exists {
		enhanced = msg
	} else {
		return core.AgentResult{}, fmt.Errorf("no enhanced data found")
	}

	// Format with LLM
	prompt := core.Prompt{
		System: "You are a formatter agent. Present information in a clear, professional, and well-structured manner.",
		User:   fmt.Sprintf("Format this response in a clear, professional manner: %v", enhanced),
	}

	response, err := a.llm.Call(ctx, prompt)
	if err != nil {
		return core.AgentResult{}, err
	}

	// Update state with final result
	outputState := core.NewState()
	outputState.Set("final_response", response.Content)
	outputState.Set("message", response.Content)

	// Print the final result
	fmt.Printf("\n📝 Final Response:\n%s\n", response.Content)

	return core.AgentResult{OutputState: outputState}, nil
}
