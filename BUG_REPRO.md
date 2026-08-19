# Bug Reproduction

## Bug
Inspection coordination publishes completion after the first worker, drops worker errors, exposes report slices, and replaces the caller context.

## Trigger
Run the five targeted inspection concurrency tests recorded in collection.json.

## Error
The baseline returns one result before slow workers finish, records zero worker errors, streams too few results, mutates published reports, and gives workers a context without cancellation.
