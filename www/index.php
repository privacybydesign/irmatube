<?php require_once "config.php"?>
<!DOCTYPE html>
<html lang="<?= $language ?>">

<head>
  <meta http-equiv="Content-Type" content="text/html; charset=UTF-8">
  <meta name="keywords" content="Yivi, YiviTube, film, privacy, security">
  <meta name="description" content="Experimental YiviTube video streaming service">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">

  <title>YiviTube - Watch movies without others noticing it!</title>

  <link href="css/mosaic.css" rel="stylesheet" type="text/css" />
  <link href="css/irmatube.css" rel="stylesheet" type="text/css" />
  <link href="node_modules/bootstrap/dist/css/bootstrap.min.css" rel="stylesheet" />

  <script src="node_modules/jquery/dist/jquery.min.js" type="text/javascript"></script>
  <script src="js/mosaic.1.0.1.min.js" type="text/javascript"></script>
  <script src="node_modules/mustache/mustache.min.js" type="text/javascript"></script>
  <script src="node_modules/bootstrap/dist/js/bootstrap.min.js" type="text/javascript"></script>
  <script src="content/movies.js" type="text/javascript"></script>

  <script src="node_modules/@privacybydesign/yivi-frontend/dist/yivi.js" type="text/javascript" async></script>

  <script id="play-ytvideo" type="text/template">
    <div class="modal fade" tabindex="-1" role="dialog" id="video_div_{{id}}">
      <div class="modal-dialog modal-lg">
        <div class="modal-content">
          <div class="modal-body">
            <iframe 
                    src="https://www.youtube-nocookie.com/embed/{{youtubeId}}"
            ></iframe>
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
        <img alt="{{title}}" src="content/covers/{{id}}.jpg">
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
          showSuccess("You are now registered for YiviTube");
      };

      yivi.newPopup({
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

      let onVerifySuccess = function (data) {
        console.log("IRMA Session data:", data);

        fetch("php/verifysession.php", {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          //cannot pass the age limit here, as the request could be tampered with

          body: JSON.stringify({
            token: data,
            videoid: videoNumber,
          }),
        })
          .then((response) => {
            if (!response.ok) {
              throw new Error("HTTP error, status = " + response.status);
            }
            return response.json();
          })
          .then((result) => {
            console.log("Result from PHP script:", result);

            if (result.success) {
              // Find the movie in IRMATubeMovies
              let movie = IRMATubeMovies.find((m) => m.id === videoNumber);
              if (!movie) {
                console.error("Movie not found for ID:", videoNumber);
                $(`#movie_${videoNumber}_player`).html(
                  "<p>Error: Movie not found.</p>"
                );
                return;
              }

              let videoData = {
                id: movie.id,
                youtubeId: result.youtubeId,
              };

              let html = Mustache.render($("#play-ytvideo").html(), videoData);
              $("body").append(html);
              $("#video_div_" + videoData.id).modal("show");
            } else {
              console.warn("Verification failed:", result.message);
              $(`#movie_${videoNumber}_player`).html(
                `<p>${result.message || "Verification failed."}</p>`
              );
            }
          })
      };

      let url = "php/session.php?type=verification";
      if (ageLimit > 0)
        url += "&age=" + ageLimit;
      url += "&" + Math.random(); // Append randomness so that IE doesn't consider it 304 not modified

      yivi.newPopup({
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
    //makes sure the youtube video stops playing when the modal is closed
    function closeMovie(videoNumber) {
      const modal = document.getElementById(`video_div_${videoNumber}`);
      if (modal) {
        const iframe = modal.querySelector('iframe');
        if (iframe) {
          iframe.src = ''; 
        }
        $(modal).modal('hide'); 
        modal.parentNode.removeChild(modal); 
      }
    }

    //uses movie.js to populate movie gallery
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
              <h4 id="registerModalLabel"> Register for YiviTube </h4>
        </div>
        <div class="modal-body">
          <p>
          You can now register for YiviTube using your Yivi identity wallet. You will get access to:
          <ol>
            <li>Eight splendid movie-trailers</li>
            <li>Automatic Yivi age verification</li>
          </ol>
          Best of all, it is totally
          </p>
          <p class="text-center">
            <h1 class="text-center">Free!</h1>
          </p>
        </div>
        <div class="modal-footer">
          <button class="btn" data-dismiss="modal" aria-hidden="true">Cancel</button>
          <button class="btn" style="background-color: darkred; color: white;" data-dismiss="modal"  aria-hidden="true" id="IRMARegister">Register using Yivi</button>
        </div>
      </div>
    </div>
  </div>

  <div class="container">
    <div id="irmaTubeHeading" class="row">
      <div class="col-md-3">
        <a href="/demo"><img src="img/YiviTube_logo.png" width="200"/></a>
      </div>
      <div class="col-md-7">
        <div id="alert_box">
        </div>
      </div>
      <div class="col-md-2">
      <button class="btn pull-right" style="background-color: darkred; color: white;" data-toggle="modal" data-target="#registerModal">Register</button>
      </div>
    </div>
    <div class="row">
      <div class="col-md-12">
        <div>
          <img src="img/arrow_red.png" id="arrow" />
          YiviTube is the privacy-friendly video-streaming service
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
