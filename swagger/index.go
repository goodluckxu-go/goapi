package swagger

const index = `<!-- HTML for static distribution bundle build -->
<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="UTF-8">
    <title>%s - Swagger UI</title>
    <link rel="stylesheet" type="text/css" href="%s` + cssSwaggerUiPath + `" />
    <link rel="stylesheet" type="text/css" href="%s` + cssIndexPath + `" />
    <link rel="icon" type="image/png" href="%s" sizes="32x32" />
  </head>

  <body>
    <div id="swagger-ui"></div>
    <script src="%s` + jsSwaggerUiBundlePath + `" charset="UTF-8"> </script>
    <script src="%s` + jsSwaggerUiStandalonePresetPath + `" charset="UTF-8"> </script>
    <script src="%s` + jsSwaggerInitializerPath + `" charset="UTF-8"> </script>
  </body>
</html>
`
