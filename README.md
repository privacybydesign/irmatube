# IRMATube

[IRMATube](https://privacybydesign.foundation/demo/irmaTube/) is a demo of an [IRMA](https://privacybydesign.foundation/irma-en) attribute issuance and verifier. It offers membership attributes to users, who then need to disclose these in order to watch a movie.

It showcases the following features:

* Issuance and verification of attributes to users.
* Selective disclosure of attributes: although an identifying membership number is issued to the user, this number is not disclosed during verification, so that IRMATube knows that a member is watching a movie, but not which member.
* Granting access to a resource only after successful IRMA attribute verification.

## Installation

* In the `www` directory:
  * `yarn` or `npm install`
  * `composer install`
* The `data` directory contains the movies and private and public keys, and should be outside your webserver's webroot. Be sure to point to it in your `config.php`, see the included example. 
* Follow the instructions in the `data` and `www/content` folders to install your movies.