# Contributing to Squirrel Services

Thank you for your interest in contributing to Squirrel Services! This document provides guidelines and instructions for contributing.

## Development Setup

### Prerequisites

- Go 1.21+ (for m03)
- Python 3.11+ (for m04)
- Docker & Docker Compose
- Make

### Local Development

```bash
# Clone the repository
git clone https://github.com/squirrelawake/squirrel-services.git
cd squirrel-services

# Start dependencies
docker-compose -f docker/docker-compose.yml up -d tidb redis

# Run m03 locally
cd m03-trajectory
go mod download
go run cmd/server/main.go

# Run m04 locally (in another terminal)
cd m04-insight-engine
python -m venv venv
source venv/bin/activate
pip install -r requirements.txt
python -m uvicorn src.main:app --reload
```

## Code Style

### Go (m03)

- Follow [Effective Go](https://golang.org/doc/effective_go.html)
- Use `gofmt` for formatting
- Run `golangci-lint` before committing
- Write tests for new functionality

### Python (m04)

- Use [Black](https://black.readthedocs.io/) for formatting
- Follow PEP 8 style guide
- Use type hints for function signatures
- Run `flake8` and `mypy` before committing

## Testing

### Running Tests

```bash
# All tests
make test

# M03 only
cd m03-trajectory && go test -v ./...

# M04 only
cd m04-insight-engine && pytest tests/ -v
```

### Test Coverage

- M03: Target >80% coverage
- M04: Target >85% coverage

## Pull Request Process

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/my-feature`)
3. Make your changes
4. Add tests for new functionality
5. Run tests and ensure they pass
6. Update documentation if needed
7. Commit with clear messages
8. Push to your fork
9. Create a Pull Request

### PR Checklist

- [ ] Tests pass
- [ ] Code follows style guidelines
- [ ] Documentation updated
- [ ] CHANGELOG.md updated
- [ ] Commit messages are clear

## Commit Message Format

```
type(scope): subject

body (optional)

footer (optional)
```

Types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation only
- `style`: Code style (formatting)
- `refactor`: Code refactoring
- `test`: Adding tests
- `chore`: Maintenance tasks

Example:
```
feat(m03): add Kalman filter for speed smoothing

Implement KalmanFilter struct with Update method for
smoothing GPS speed measurements.

Closes #123
```

## Reporting Issues

When reporting bugs, please include:

- Go/Python version
- Operating system
- Steps to reproduce
- Expected behavior
- Actual behavior
- Error messages/logs

## Questions?

Feel free to open an issue for questions or join our discussions.

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
