# Bug Reproduction

## Bug
Manifest snapshots share mutable storage with the live store and publisher. Historical seals change after later writes, and concurrent access reports a data race.

## Trigger
Run the four targeted manifest snapshot, concurrent view, revision publication, and publisher isolation tests recorded in collection.json.

## Error
The baseline reports `WARNING: DATA RACE`, `historical snapshot changed`, and mismatched published revision values.
