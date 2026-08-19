# Bug Reproduction

## Bug
The occupancy ledger exposes its internal map and publishes revision separately from slot changes; a move performs two locked mutations, so callers can corrupt the ledger and observers see torn snapshots.

## Trigger
Run the four targeted occupancy snapshot, concurrent view, single-revision move, and cache isolation tests recorded in collection.json.

## Error
The baseline lets callers delete ledger keys through the snapshot, reports a torn hazardous count under concurrency, publishes two revisions per move, and exposes cache maps.
