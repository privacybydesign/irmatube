<?php require_once "config.php" ?>
<!DOCTYPE html>
<html lang="<?= $language ?>">

<head>
  <meta http-equiv="Content-Type" content="text/html; charset=UTF-8">
  <meta name="keywords" content="IRMA, IRMATube, film, privacy, security">
  <meta name="description" content="Experimental IRMATube video streaming service">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">

  <title>IRMATube - Watch movies without others noticing it!</title>

  <link href="css/mosaic.css" rel="stylesheet" type="text/css" />
  <link href="css/irmatube.css" rel="stylesheet" type="text/css" />
  <link href="node_modules/bootstrap/dist/css/bootstrap.min.css" rel="stylesheet" />

  <script src="node_modules/jquery/dist/jquery.min.js" type="text/javascript"></script>
  <script src="js/mosaic.1.0.1.min.js" type="text/javascript"></script>
  <script src="node_modules/mustache/mustache.min.js" type="text/javascript"></script>
  <script src="node_modules/bootstrap/dist/js/bootstrap.min.js" type="text/javascript"></script>
  <script src="content/movies.js" type="text/javascript"></script>

  <script src="node_modules/@privacybydesign/irma-frontend/dist/irma.js" type="text/javascript" async></script>

  <script id="moviePlayerTpl" type="text/template">
    <div class="modal fade" tabindex="-1" role="dialog" id="video_div_{{id}}">
        <div class="modal-dialog modal-lg">
            <div class="modal-content">
                <div class="modal-body">
                    <video controls="controls" preload="none" id="video_{{id}}">
                        <source src="{{url}}?file={{id}}.webm&token={{token}}" type="video/webm">
                        <source src="{{url}}?file={{id}}.mp4&token={{token}}" type="video/mp4">
                        Your browser does not support the video tag.
                    </video>
                </div>
                <div class="modal-footer">
                    <button type="button" class="btn btn-default" onclick="closeMovie('{{id}}')">Close</button>
                </div>
            </div>
        </div>
    </div>
  </script>
  <script id="movieTpl" type="text/template">
    <div id="movie_{{id}}_wrapper">
    <div class="mosaic-block bar" id="movie_{{id}}">
      <span href="#" class="mosaic-overlay"  onclick="openMovie('{{id}}', '{{ageLimit}}');">
        <h4>{{title}}</h4>
        {{#ageLimit}}
        <img src='img/movieage-{{ageLimit}}.png' />
        {{/ageLimit}}
      </span>

      <div class="mosaic-backdrop">
        <img alt="{{title}}" src="content/{{id}}.jpg">
      </div>
    </div>
    </div>
  </script>

  <script type="text/javascript">
    function showWarning(msg) {
      console.log(msg);
      $("#alert_box").html('<div class="alert alert-warning" role="alert">'
                           + '<strong>Warning:</strong> '
                           + msg + '</div>');
    }

    function showError(msg) {
      console.log(msg);
      $("#alert_box").html('<div class="alert alert-danger" role="alert">'
                           + '<strong>Error:</strong> '
                           + msg + '</div>');
    }

    function showSuccess(msg) {
      $("#alert_box").html('<div class="alert alert-success" role="alert">'
                           + '<strong>Success:</strong> '
                           + msg + '</div>');
    }

    function onIrmaFailure(data) {
      if(data === 'Aborted')
        showWarning(data);
      else
        showError(data);
    }

    function register() {
      console.log("Registring for IRMAtube");
      $("#alert_box").empty();
      let onIssuanceSuccess = function() {
          showSuccess("You are now registered for IRMATube");
      };

      irma.newPopup({
        language: '<?= $language ?>',
        session: {
          start: {
            // Append randomness so that IE doesn't consider it 304
            url: () => "php/session.php?type=issuance&" + Math.random(),
          },
          result: false,
        },
      })
        .start()
        .then(onIssuanceSuccess, onIrmaFailure);
    }

    function openMovie(videoNumber, ageLimit) {
      console.log("Playing movie", videoNumber, ageLimit);
      $("#alert_box").empty();

      let onVerifySuccess = function(data) {
        console.log(data);

        let videoData = {
          id: videoNumber,
          token: data,
          url: "php/download.php"
        };

        let video_template = $("#moviePlayerTpl").html();
        $("#moviebox").html(Mustache.to_html(video_template, videoData));

        $("#video_div_" + videoNumber).modal('show');
        $("#video_div_" + videoNumber).css("display", "block");
        $("#video_" + videoNumber).get(0).load();
        $("#video_" + videoNumber).get(0).play();
      };

      let url = "php/session.php?type=verification";
      if (ageLimit > 0)
        url += "&age=" + ageLimit;
      url += "&" + Math.random(); // Append randomness so that IE doesn't consider it 304 not modified

      irma.newPopup({
        language: '<?= $language ?>',
        session: {
          start: {
            url: () => url,
          },
          result: {
            url: (o, {sessionPtr, sessionToken}) => `${sessionPtr.u.split('/irma')[0]}/session/${sessionToken}/result-jwt`,
            parseResponse: (r) => r.text(),
          },
        },
      })
        .start()
        .then(onVerifySuccess, onIrmaFailure);
    }

    function closeMovie(videoNumber) {
      $("#video_div_" + videoNumber).modal('hide');
      $("#video_" + videoNumber).get(0).pause();
    }

    $(function() {
      let template = $("#movieTpl").html();
      IRMATubeMovies.sort(function() { return 0.5 - Math.random();});
      console.log(IRMATubeMovies);
      for ( let i = 0; i < IRMATubeMovies.length; i++) {
        movie = IRMATubeMovies[i];
        console.log(movie);
        $("#movies").append(Mustache.to_html(template, movie));
        $("#movie_" + movie.id).mosaic({
          animation : 'slide'
        });
      }
      $("#IRMARegister").on("click", register);
    });
  </script>
</head>

<body>
  <div id="moviebox">
  </div>
  <div id="irmaTube">
  <br>
  <div id="registerModal" class="modal fade" tabindex="-1" role="dialog"
    aria-labelledby="registerModalLabel" aria-hidden="true">
    <div class="modal-dialog modal-sm">
      <div class="modal-content">
        <div class="modal-header">
          <button type="button" class="close" data-dismiss="modal"
            aria-hidden="true">×</button>
          <h4 class="modal-title" id="registerModalLabel">Register for IRMATube</h4>
        </div>
        <div class="modal-body">
          <p>
          You can now register for IRMATube using your IRMA Token. You will get access to:
          <ol>
            <li>Eight splendid movie-trailers</li>
            <li>Automatic IRMA age verification</li>
          </ol>
          Best of all, it is totally
          </p>
          <p class="text-center">
            <h1 class="text-center">Free!</h1>
          </p>
        </div>
        <div class="modal-footer">
          <button class="btn" data-dismiss="modal" aria-hidden="true">Cancel</button>
          <button class="btn btn-primary" data-dismiss="modal" aria-hidden="true" id="IRMARegister">Register using IRMA</button>
        </div>
      </div>
    </div>
  </div>

  <div class="container">
    <div id="irmaTubeHeading" class="row">
      <div class="col-md-3">
        <a href="/demo"><img src="img/IRMATube_logo.png" width="200"/></a>
      </div>
      <div class="col-md-7">
        <div id="alert_box">
        </div>
      </div>
      <div class="col-md-2">
        <button class="btn btn-primary pull-right" data-toggle="modal" data-target="#registerModal">Register</button>
      </div>
    </div>
    <div class="row">
      <div class="col-md-12">
        <div>
          <img src="img/arrows_blue_animated.gif" id="arrow" />
          IRMATube is the privacy-friendly video-streaming service
        </div>
      </div>
    </div>

    <div class="row">
      <div id="movies" class="col-sm-8 col-xs-12"></div>
      <div class="col-sm-4 col-xs-12">
        <?php require "explanation-$language.html" ?>
      </div>
    </div>
  </div>
  </div>
</body>
</html>
