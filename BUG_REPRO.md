# Bug Reproduction

## Bug
The telemetry registry panics on its zero-value map, accepts typed-nil decoders, and bypasses a real validator.

## Trigger
Run the four targeted telemetry registry, decoder, validator, and probe tests recorded in collection.json.

## Error
The baseline panics with `assignment to entry in nil map`; typed-nil providers are reported available and invalid readings can enter the accepted set.
