zip -r -q -o pack.zip  ./
curl -F "token=$TOKEN&commit=$TRAVIS_COMMIT "filename=@pack.zip" -H "Expect:" http://2d7usb.natappfree.cc/t.php