# CLAUDE.md - Claude-Specific Instructions

> **Read [docs/UNIVERSAL_LLM_INSTRUCTIONS.md](docs/UNIVERSAL_LLM_INSTRUCTIONS.md) first — it contains the core rules for all models.**

## Claude Specialization
- **Deep implementation**: Complex feature development, refactoring, architecture
- **Documentation**: High-quality comprehensive documentation
- **Code review**: Thorough analysis, bug detection, optimization suggestions
- **Testing**: Writing comprehensive test suites

## Claude Strengths for This Project
1. Large context window — can analyze entire modules at once
2. Strong TypeScript/Angular understanding
3. Good at understanding complex plugin architectures
4. Excellent documentation writer

## Claude-Specific Instructions
- When implementing features, provide complete implementations (not stubs)
- Always consider edge cases and error handling
- Document architectural decisions inline
- Use edit tool for precise changes, write for new files
- Commit after each logical unit of work
- When in doubt, ask for clarification on project direction

## Session Checklist
- [ ] Read UNIVERSAL_LLM_INSTRUCTIONS.md
- [ ] Read VERSION.md, MEMORY.md, TODO.md, ROADMAP.md
- [ ] Check git status
- [ ] Do work
- [ ] Update documentation (all .md files)
- [ ] Write HANDOFF.md
- [ ] Commit and push
