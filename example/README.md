# Usage Example

## Build & Run

Just call in the _webapp-backend_ folder:

```bash
go get ./...
go build
```

then you can copy the executable anywhere you like and run it from there. All static files are embedded into that executable.

## Modify Frontend

In the _webapp-frontend_ folder follow these steps:

```bash
npm i # Restore all npm package
npm start # Run the development server
# Now make any changes in the webapp frontend you like (ex. add static files, change components, css or js files etc...)

# Next build the production files
npm run build
```

Now change into the _webapp-backend_ folder and execute the following commands:

```bash
go generate
go get ./... # Only if not already done before
go build
```

Again copy the executable anywhere you like and then run it from there.
