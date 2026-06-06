# Claude Code Analysis & Feature Mapping

## Overview
Claude Code is an agentic coding tool that provides a structured workflow for feature development, codebase exploration, and quality review directly from the terminal.

## Key Features & Workflow (Feature Dev Plugin)
The most valuable part of Claude Code is its **7-Phase Workflow** for building features:

1.  **Discovery**: Identify requirements and constraints.
2.  **Exploration**: Launch parallel `code-explorer` agents to map architecture and find similar features.
3.  **Clarification**: Address ambiguities before starting design.
4.  **Architecture Design**: Generate multiple approaches (Minimal, Clean, Pragmatic) and present trade-offs.
5.  **Implementation**: Execute code changes based on the chosen design.
6.  **Quality Review**: Launch parallel `code-reviewer` agents (Simplicity, Correctness, Conventions).
7.  **Summary**: Document changes and suggest next steps.

## Specialized Agents
- `code-explorer`: Traces execution paths and data flow.
- `code-architect`: Designs implementation blueprints.
- `code-reviewer`: Identifies bugs and convention violations.

## Integration Plan for Tabby Go Harness

### Backend (Go)
1.  **Task Graph Orchestration**:
    - Update `tabby-go/pkg/agent` to support "Phase-based" tasks.
    - Implement a `WorkflowManager` that can track a task through Discovery -> Design -> Review.
2.  **Parallel Agent Workers**:
    - Allow the Go backend to spawn multiple concurrent LLM requests (using the `ai.go` client) to simulate parallel "agents" (explorer, architect, reviewer).
3.  **Context Buffering**:
    - Implement a system to persist "Codebase Discovery" findings across session phases.

### Frontend (Angular)
1.  **Workflow UI**:
    - Build a dedicated "Workflow Progress" component to visualize the 7 phases.
    - Implement a "Clarification Modal" that blocks the workflow until user input is received (matching Claude Code's Phase 3).
2.  **Choice Selector**:
    - Create a UI for users to pick between multiple architecture approaches (Matching Phase 4).

## Status
- **Analysis**: Complete.
- **Implementation**: Partially covered by existing `agent.go` task status. Needs explicit Workflow/Phase logic.
- **Redundancy**: Submodule `claude-code` is primarily a collection of workflow definitions and scripts. The core logic can be reimplemented in Go.
