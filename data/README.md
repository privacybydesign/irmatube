Data directory
--------------

Add the following to this directory:

 * `pk.pem` - The public key of the IRMA server you are using
 * `videos/` - A directory containing for each movie:
   * A `.webm` version
   * An `.mp4` version
   * An `.access` file containing the minimum age, or 0 if there is no age limit for this movie.

If you change the names of these files, or move the location of this folder relative to the www/ directory, be sure to update `www/config.php` and `www/content/movies.js`.