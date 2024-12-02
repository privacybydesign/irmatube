Data directory
--------------

Add the following to this directory:

 * `pk.pem` - The public key of the IRMA server you are using
 * `videos/` - A directory containing for each movie:
  * A `.json` containing age restriction and youtube ID for embedding the video


If you change the names of these files, or move the location of this folder relative to the www/ directory, be sure to update `www/config.php` and `www/content/movies.js`.