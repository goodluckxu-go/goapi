package swagger

const index = `<!-- HTML for static distribution bundle build -->
<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="UTF-8">
    <title>%s - Swagger UI</title>
    <link rel="stylesheet" type="text/css" href="` + cssSwaggerUiPath + `" />
    <link rel="stylesheet" type="text/css" href="` + cssIndexPath + `" />%s
    <link rel="icon" type="image/png" href="` + favicon + `" sizes="32x32" />
  </head>

  <body>
    <div id="swagger-ui"></div>
    <script src="` + jsSwaggerUiBundlePath + `" charset="UTF-8"> </script>
    <script src="` + jsSwaggerUiStandalonePresetPath + `" charset="UTF-8"> </script>
    <script src="` + jsSwaggerInitializerPath + `" charset="UTF-8"> </script>
  </body>
</html>
`
