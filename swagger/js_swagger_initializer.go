package swagger

const jsSwaggerInitializerPath = "/swagger-initializer.js"

const openapiPath = "/openapi.json"

const jsSwaggerInitializer = `window.onload = function() {
  //<editor-fold desc="Changeable Configuration Block">

  // the following lines will be replaced by docker/configurator, when it runs in a docker-container
  window.ui = SwaggerUIBundle({
    url: "%s` + openapiPath + `",
    dom_id: '#swagger-ui',
	docExpansion: '%s',
    deepLinking: %v,
    presets: [
      SwaggerUIBundle.presets.apis,
      SwaggerUIStandalonePreset
    ],
    plugins: [
      SwaggerUIBundle.plugins.DownloadUrl
    ],
    layout: "BaseLayout"
  });

  //</editor-fold>
};
`
