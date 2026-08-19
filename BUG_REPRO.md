# Bug Reproduction

## Bug
Hazard-wagon filtering compacts the input slice in place; routing rewrites and saved reservations share the same backing array, so history plans and the original consist are corrupted by later edits.

## Trigger
Run the four targeted consist filtering, plan independence, reservation, and export tests recorded in collection.json.

## Error
The baseline mutates the input after `HazardOnly`, lets `Routed` edits leak into `Original`/`Hazard`, corrupts stored reservations, and exposes plan storage through `ExportPlan`.
