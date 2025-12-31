# Contributing to YFlow

Thank you for your interest in contributing to YFlow! This document provides guidelines and instructions for contributing.

## Getting Started

### Prerequisites

- **Go** 1.21+ (for backend)
- **Node.js** 18+ (for frontend)
- **pnpm** 8+ (for frontend)
- **Bun** 1.0+ (for CLI)
- **MySQL** 8.0+
- **Redis** 7.0+

### Development Setup

1. **Fork the repository** on GitHub

2. **Clone your fork**:
   ```bash
   git clone https://github.com/YOUR_USERNAME/yflow.git
   cd yflow
   ```

3. **Set up upstream remote**:
   ```bash
   git remote add upstream https://github.com/yflow-io/yflow.git
   ```

4. **Install dependencies**:
   ```bash
   # Backend
   cd admin-backend && go mod tidy

   # Frontend
   cd admin-frontend && pnpm install

   # CLI
   cd cli && bun install
   ```

5. **Start development services**:
   ```bash
   docker compose up -d
   ```

## Development Workflow

### 1. Create a Branch

```bash
git checkout main
git fetch upstream
git merge upstream/main
git checkout -b feature/your-feature-name
```

**Branch naming conventions:**
- `feature/` - New features
- `fix/` - Bug fixes
- `docs/` - Documentation changes
- `refactor/` - Code refactoring
- `test/` - Test additions/changes

### 2. Make Changes

Follow the code standards defined in CLAUDE.md:

**Backend (Go):**
- Follow Clean Architecture patterns
- Use Uber FX for dependency injection
- Add unit tests for new functionality

**Frontend (Vue 3):**
- Use TypeScript
- Follow component patterns in existing code
- Run `pnpm lint` before committing

**CLI (Bun):**
- Use TypeScript
- Follow existing command structure

### 3. Commit Your Changes

We follow [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

**Types:**
- `feat` - A new feature
- `fix` - A bug fix
- `docs` - Documentation only changes
- `style` - Changes that do not affect the meaning of the code (white-space, formatting, etc)
- `refactor` - A code change that neither fixes a bug nor adds a feature
- `perf` - A code change that improves performance
- `test` - Adding missing tests or correcting existing tests
- `chore` - Changes to the build process or auxiliary tools

**Examples:**
```
feat(user): add user avatar upload
fix(auth): resolve token refresh issue
docs(readme): update installation instructions
refactor(service): simplify cache invalidation logic
```

### 4. Run Tests and Lint

```bash
# Backend
cd admin-backend
go test ./...           # Run all tests
go vet ./...            # Go vet
# golangci-lint run     # If configured

# Frontend
cd admin-frontend
pnpm type-check         # Type checking
pnpm lint               # Lint
pnpm test:unit          # Run tests
```

### 5. Push and Create PR

```bash
git push origin feature/your-feature-name
```

Then open a Pull Request on GitHub.

## Pull Request Guidelines

### PR Description

Include the following in your PR:

- **What** changes were made
- **Why** these changes were necessary
- **How** the changes were implemented
- **Screenshots** (for UI changes)
- **Test plan** - How you tested the changes

### PR Requirements

- [ ] Code follows project's coding standards
- [ ] Tests pass (if applicable)
- [ ] No linting errors
- [ ] Documentation updated (if needed)
- [ ] Commit messages follow conventional format
- [ ] PR description is clear and complete

### Review Process

1. All PRs require at least one maintainer approval
2. Maintainers may request changes or clarifications
3. Once approved, a maintainer will merge your PR
4. Please be responsive to review feedback

## Code Style Guidelines

### Go Backend

- Use `gofmt` for formatting
- Follow [Effective Go](https://go.dev/doc/effective_go)
- Use meaningful variable names
- Add comments for exported functions/types

### TypeScript/Vue Frontend

- Use ESLint + Prettier (configured in project)
- Follow existing component patterns
- Use TypeScript for all new code
- Use Composition API with `<script setup>`

## Security Considerations

- **Never** commit secrets or credentials
- Use environment variables for sensitive data
- Follow the security guidelines in backend README
- Report security vulnerabilities privately to maintainers

## Community

- Be respectful and constructive in discussions
- Help others by answering questions
- Follow the project's Code of Conduct

## Questions?

If you have questions, feel free to:

- Open an issue with the `question` tag
- Check existing documentation in `/docs`
- Review the API documentation at `/swagger` when running locally

---

**Thank you for contributing to YFlow!**
