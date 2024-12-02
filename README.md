# YiviTube

YiviTube is a demo attribute issuance and verifier. It offers membership attributes to users, who then need to disclose these in order to watch a movie.

It showcases the following features:

* Issuance and verification of attributes to users.
* Selective disclosure of attributes: although an identifying membership number is issued to the user, this number is not disclosed during verification, so that IRMATube knows that a member is watching a movie, but not which member.
* Granting access to a resource only after successful IRMA attribute verification.

Note that the movies themselves are not included in this repository.

## Installation

### Without Docker

* In the `www` directory:
  * `yarn` or `npm install`
  * `composer install`
* The `data` directory contains the movies and private and public keys, and should be outside your webserver's webroot. Be sure to point to it in your `config.php`, see the included example. 
* Follow the instructions in the `data` and `www/content` folders to install your movies. There are already example movies set up in the config file, movies.js and data/videos directory with covers in content/covers. You can customize and change these as you wish.
* Lastly, in the www directory, run a php server (e.g php -S localhost:8080). You will also need to run an irma server.

### With Docker

You can use the `docker compose up --build` to build and test the project in one go. 

