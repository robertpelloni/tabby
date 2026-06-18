# Just-Every-Code Analysis

## Overview
`just-every-code` is a fast, local coding agent for the terminal. It is a community-driven fork of `openai/codex` focused on real developer ergonomics.

## Key Features Discovered
- **Multi-Agent Commands**:
  - `/plan`: Coordinates multiple agents to review a task and create a consolidated plan.
  - `/solve`: Starts a race between models to solve complex problems, taking the fastest optimal solution.
  - `/code`: Creates multiple worktrees to implement an optimal solution based on consensus.
- **Browser Integration**:
  - `/chrome`: Connects the agent to an external Chrome browser (via CDP) to drive automation or context gathering.
  - `/browser`: Connects the agent to an internal headless browser.
- **Auto Drive**:
  - Handoff capability for multi-step tasks where the agent coordinates approvals and sub-tasks without user intervention.
- **MCP (Model Context Protocol)**:
  - Natively supports MCP via configuration (`~/.code/config.toml`) for filesystem access, database queries, and external tool execution.
