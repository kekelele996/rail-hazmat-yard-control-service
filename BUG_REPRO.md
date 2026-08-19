# Bug Reproduction

## Bug
Incident batch processing defers every resource close to the end of the function, the transaction commit overwrites the named return value, and commit failures are committed before rollback can happen.

## Trigger
Run the four targeted incident resource, business-error, commit-failure, and cleanup tests recorded in collection.json.

## Error
The baseline exhausts the resource limit on a 20-item batch, loses the original business error, commits before failing, and drops cleanup errors.
