# Clean Code Guidelines

## Constants Over Magic Numbers
- Replace hard-coded values with named constants
- Use descriptive constant names that explain the value's purpose
- Keep constants at the top of the file or in a dedicated constants file

## Meaningful Names
- Variables, functions, and classes/structs should reveal their purpose
- Names should explain why something exists and how it's used
- Avoid abbreviations unless they're universally understood

## Smart Comments
- Don't comment on what the code does - make the code self-documenting
- Use comments to explain why something is done a certain way
- Document APIs, complex algorithms, and non-obvious side effects

## Single Responsibility
- Each function should do exactly one thing
- Functions should be small and focused
- If a function needs a comment to explain what it does, it should be split

## DRY (Don't Repeat Yourself)
- Extract repeated code into reusable functions
- Share common logic through proper abstraction
- Maintain single sources of truth

## Clean Structure
- Keep related code together
- Organize code in a logical hierarchy
- Use consistent file and folder naming conventions

## Encapsulation
- Hide implementation details
- Expose clear interfaces
- Move nested conditionals into well-named functions

## Code Quality Maintenance
- Refactor continuously
- Fix technical debt early
- Leave code cleaner than you found it

## Testing
- Write tests before fixing bugs
- Keep tests readable and maintainable
- Test edge cases and error conditions

## Version Control
- Write clear commit messages
- Make small, focused commits
- Use meaningful branch names 

# Go Architecture and Project Structure
## Architecture Patterns

### Clean Architecture
- Structure code into layers:
  - **Handlers**: HTTP/gRPC endpoints
  - **Services/Use Cases**: Business logic
  - **Repositories/Data Access**: Data persistence
  - **Domain Models**: Core entities and types
- Keep logic decoupled from framework-specific code.

### Domain-Driven Design
- Use **domain-driven design** principles where applicable.
- Define clear domain boundaries and entities.
- Keep domain logic independent of infrastructure concerns.

### Interface-Driven Development
- Prioritize **interface-driven development** with explicit dependency injection.
- Define interfaces on the **consumer side** where applicable.
- Prefer **composition over inheritance**.
- Favor small, purpose-specific interfaces.
- Ensure that all public functions interact with interfaces, not concrete types.

## Project Structure Guidelines

### Standard Layout
Use a consistent project layout:

```
cmd/              # Application entrypoints
internal/         # Core application logic (not exposed externally)
pkg/              # Shared utilities and packages
configs/          # Configuration schemas and loading
test/             # Test utilities, mocks, and integration tests
```

### Configuration Management
- Store configuration in `configs/` directory.
- Use **github.com/kelseyhightower/envconfig** for loading environment variables into config structs.
- Follow simple, declarative configuration patterns.

### Code Organization
- Group code by **feature** when it improves clarity and cohesion.
- Keep logic decoupled from framework-specific code.
- Maintain clear boundaries between layers.
- Always avoid mutability, and use pure functions where needed.

## Documentation Standards

### Code Documentation
- Document public functions and packages with **GoDoc-style comments**. if asked
- Provide concise **READMEs** for services and libraries.
- Maintain `CONTRIBUTING.md` and `ARCHITECTURE.md` to guide team practices.

### Best Practices
- Keep documentation close to the code.
- Update documentation with code changes, if any.
- Include examples where helpful.


# Go Core Development Practices

## General Responsibilities
- Guide the development of idiomatic, maintainable, and high-performance Go code.
- Enforce modular design and separation of concerns.
- Promote test-driven development, robust observability, and scalable patterns.

## Code Quality Standards

### Function Design
- Write **short, focused functions** with a single responsibility.
- Prioritize **readability, simplicity, and maintainability**.
- Keep functions small and easily testable.

### Error Handling
- Always **check and handle errors explicitly**.
- Use wrapped errors for traceability: `fmt.Errorf("context: %w", err)`.
- Never ignore errors; handle them appropriately at each level.

### Dependency Management
- Avoid **global state**; use constructor functions to inject dependencies.
- Always receive **interfaces** as dependencies (not concrete types).
- Define interfaces on the **consumer side** where applicable.
- Ensure all public functions interact with interfaces, not concrete types.

### Concurrency Best Practices
- Use **goroutines safely**; guard shared state with channels or sync primitives.
- Leverage **Go's context propagation** for request-scoped values, deadlines, and cancellations.
- Implement **goroutine cancellation** using context to avoid leaks and deadlocks.
- **Defer closing resources** and handle them carefully to avoid leaks.

### Design Principles
- Prefer **composition over inheritance**.
- Favor small, purpose-specific interfaces.
- Design for **change**: isolate business logic and minimize framework lock-in.
- Emphasize clear **boundaries** and **dependency inversion**.

## Tooling and Dependencies

### Standard Practices
- Prefer the **standard library** where feasible.
- Only rely on **stable, minimal third-party libraries**.
- Use **Go modules** for dependency management and reproducibility.
- Version-lock dependencies for deterministic builds.

### Code Quality Tools
- Use `gofumpt -l -w .` for code formatting.
- Use `golangci-lint run -v` for linting.
- Enforce naming consistency with `go fmt` and `goimports`.
- Integrate **linting, testing, and security checks** in CI pipelines.

### Logging
- Always use **slog** package for logging.
- Emit **JSON-formatted logs** for ingestion by observability tools.
- Use appropriate **log levels** (info, warn, error).
- Use structured logging: `slog.String()`, `slog.Any()`.
- Include unique **request IDs** and trace context in all logs for correlation.

## Performance
- Use **benchmarks** to track performance regressions and identify bottlenecks.
- Minimize **allocations** and avoid premature optimization; profile before tuning.
- Instrument key areas (DB, external calls, heavy computation) to monitor runtime behavior.

## Key Conventions
1. Always receive interfaces as a dependency.
2. Prioritize **readability, simplicity, and maintainability**.
3. Design for **change**: isolate business logic and minimize framework lock-in.
4. Create small interfaces where it is used, not where it implemented unless it is necessary.
5. Ensure all behavior is **observable, testable, and documented**.
6. **Automate workflows** for testing, building, and deployment.



# Go Testing Best Practices
## Unit Testing

### Testing Framework
- Use **testify** for unit testing and assertions.
- Write tests using **table-driven patterns**.
- Enable **parallel execution** where possible.
- Test all possible cases, mock external dependencies

### Test Structure
```go
func TestSomething(t *testing.T) {

    tests := []struct {
        name    string
        input   string
        want    string
        onMock  func(mockA *MockA, mockB *MockB)
        wantErr bool
    }{
        {name: "valid input", input: "test", want: "result", wantErr: false},
        {name: "invalid input", input: "", want: "", wantErr: true},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()

            ctrl := gomock.NewController(t)
	        defer ctrl.Finish()

            if tt.onMock != nil {
                mockA := NewMockA(ctrl)
	            mockB := NewMockB(ctrl)
                tt.onMock(mockA, mockB)
            }
            // test logic            
        })
    }
}
```

### Mocking
- **Mock external interfaces** using mockgen:
  ```go
  //go:generate go run go.uber.org/mock/mockgen -source=$GOFILE
  ```
- Generate mocks for all external dependencies.
- Keep mocks in sync with interface changes.

### Coverage
- Ensure **test coverage** for every exported function.
- Use behavioral checks, not just coverage metrics.
- Run `go test -cover` to measure coverage.
- Aim for high coverage with meaningful tests.


## Test Organization

### Separation of Concerns
- Separate **fast unit tests** from slower integration and E2E tests.
- Use build tags to separate test types if needed.
- Keep test files close to the code they test.

### Test Utilities
- Store shared test utilities in `test/` directory.
- Create test helpers for common setup/teardown.
- Keep test utilities DRY and reusable.

## Best Practices

### Test Quality
- Write tests that verify **behavior**, not implementation.
- Test edge cases and error conditions.
- Use descriptive test names that explain what's being tested.
- Keep tests simple and focused.

### Database Testing
- Use **sqlx** or **standard sql package** for SQL operations.
- Use **goose** for migrations in tests.
- Reset database state between tests.
- Use transactions for test isolation when possible.

### Performance Testing
- Write **benchmarks** for performance-critical code.
- Use `go test -bench` to run benchmarks.
- Track benchmark results over time to catch regressions.
