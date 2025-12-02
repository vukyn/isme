# Rainy Project - Cursor Rules

This directory contains focused, composable rules for the Rainy Go API project. Each rule file is written in MDC format and follows best practices for maintainability and reusability.

## Rule Files

### 1. `architecture.mdc` (Always Applied)
- **Type**: Always
- **Scope**: Clean Architecture patterns and project structure
- **Globs**: `internal/**/*`, `cmd/**/*`, `pkg/**/*`
- **Purpose**: Defines the overall architecture, project structure, and layer organization

### 2. `coding-conventions.mdc` (Always Applied)
- **Type**: Always
- **Scope**: Coding conventions, naming patterns, and style guidelines
- **Globs**: `**/*.go`
- **Purpose**: Establishes consistent coding standards, naming conventions, and Go best practices

### 3. `domain-patterns.mdc` (Auto Attached)
- **Type**: Auto Attached
- **Scope**: Domain-driven design patterns and layer implementations
- **Globs**: `internal/domains/**/*`
- **Purpose**: Provides patterns for entities, repositories, usecases, handlers, and models

### 4. `dependency-injection.mdc` (Auto Attached)
- **Type**: Auto Attached
- **Scope**: Dependency injection patterns and configuration management
- **Globs**: `internal/di/**/*`, `internal/config/**/*`, `internal/middlewares/**/*`
- **Purpose**: Defines DI patterns, configuration management, and container usage

### 5. `technology-stack.mdc` (Agent Requested)
- **Type**: Agent Requested
- **Scope**: Technology stack, libraries, and framework usage patterns
- **Purpose**: Documents the technology stack, library usage patterns, and framework-specific implementations

### 6. `error-handling.mdc` (Auto Attached)
- **Type**: Auto Attached
- **Scope**: Error handling patterns, custom exceptions, and logging strategies
- **Globs**: `**/exceptions/**/*`, `**/handlers/**/*`, `pkg/http/**/*`
- **Purpose**: Defines error handling patterns, custom exceptions, and logging strategies

## Rule Types

- **Always**: Always included in model context
- **Auto Attached**: Included when files matching glob patterns are referenced
- **Agent Requested**: Available to AI, which decides whether to include it
- **Manual**: Only included when explicitly mentioned using @ruleName

## Best Practices Applied

1. **Focused**: Each rule addresses a specific concern
2. **Actionable**: Rules provide concrete examples and patterns
3. **Scoped**: Rules are limited to relevant file patterns
4. **Under 500 lines**: Each rule file is concise and focused
5. **Composable**: Rules can be combined and reused
6. **Concrete Examples**: Rules include real code examples from the project
7. **Clear Documentation**: Rules are written like internal documentation

## Usage

These rules will automatically apply based on the file patterns you're working with. The AI will have access to relevant rules based on the context of your work, ensuring consistent code generation and adherence to project standards.
