# OpenAPI Contract

`openapi/openapi.yaml` is the contract entrypoint. Its referenced files under
`openapi/paths/` and `openapi/components/` are the editable source files.

Do not edit these generated files directly:

- `openapi/bundle.yaml` is the self-contained contract served by the API and
  consumed by Scalar.
- `web/src/api/schema.gen.ts` contains the generated TypeScript types.

## Updating the Contract

1. Change the Go handler and its tests when HTTP behavior changes.
2. Update the matching path or component source file.
3. Run `pnpm contract:generate` to rebuild the bundle and TypeScript types.
4. Run `pnpm contract:check` before committing.

Keep the root file for API metadata and `$ref` wiring. Put endpoint behavior in
`paths/` and reusable request, response, and schema definitions in
`components/`.

The backend remains the source of truth for implemented behavior. The OpenAPI
contract must not document routes or response shapes that the server does not
currently provide.
