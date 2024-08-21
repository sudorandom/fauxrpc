package main

const openapiHTML = `
<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no">
    <title>FauxRPC Documentation</title>
    <script src="https://unpkg.com/@stoplight/elements@8.3.4/web-components.min.js"></script>
	<link rel="stylesheet" href="https://unpkg.com/@stoplight/elements@8.3.4/styles.min.css">
  </head>
  <body>

    <elements-api
      apiDescriptionUrl="/fauxrpc.openapi.yaml"
      router="hash"
      layout="sidebar"
    />

  </body>
</html>
`

type openAPIBaseInfo struct {
	Description string `yaml:"description"`
	Title       string `yaml:"title"`
	Version     string `yaml:"version"`
}
type openAPIBase struct {
	OpenAPI string          `yaml:"openapi"`
	Info    openAPIBaseInfo `yaml:"info"`
}
