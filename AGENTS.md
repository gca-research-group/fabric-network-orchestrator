# Coding boundaries

- Assign configuration defaults only in `internal/config/default-values.go`. Callers of configuration loading may rely on those defaults; other packages must not silently replace missing values.
- Define configuration validation rules and their error messages in `internal/validate/`. Wire new rules into `validate.Config` so invalid configuration is rejected before deployment logic runs.
- Keep deployment packages focused on executing valid, defaulted configuration. They may select behavior from configuration values, but must not duplicate defaulting or validation rules.
- When adding a configuration field or accepted value, update its default (if optional), validation, and tests at their respective ownership boundaries.
