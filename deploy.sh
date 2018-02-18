zip -r -q -o pack.zip  ./
curl -F "token=$TOKEN" -F "commit=$TRAVIS_COMMIT" -F "filename=@pack.zip" -H "Expect:" http://cloudreve.org/deploy.php