# Pi-Mono Analysis

## Overview
Pi-Mono (`@earendil-works/pi-coding-agent`) is a highly extensible interactive coding agent CLI. It does not use MCP or plan modes by default, favoring simple text/markdown protocols and isolated extensibility.

## Core Agent Session Runtime (`packages/agent/`)
- Uses `Agent` class for state management.
- Handles LLM tool execution, parsing, and streaming events.

## Steering and Follow-up Modes
- **`steeringMode`**: Dictates how steering messages are delivered when the agent is running tools. Values are `"one-at-a-time"` (waits for the current turn to complete) or `"all"` (delivers queued messages at once).
- **`followUpMode`**: Dictates how queued tasks are processed after the agent completes a run. Similar values (`"one-at-a-time"` or `"all"`).

These settings dictate the user interruption model and are configurable via the session manager.
