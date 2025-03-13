> At first I integrated an nlp api by google in order to extract entities

https://go-firebird.fly.dev/api/testing/entity?mock=t

But this is kind of useless if we were not able to classify it on bluesky data
https://go-firebird.fly.dev/api/testing/entity



> I then added sentiment data to the document: 

https://go-firebird.fly.dev/api/testing/sentiment?mock=t

We can see the results on some live data
https://go-firebird.fly.dev/api/testing/sentiment



> Next I made the classification route to test the model that was deployed on live blue sky data

https://go-firebird.fly.dev/api/testing/classification?m=t
https://go-firebird.fly.dev/api/testing/classification

> Added a route for geocoding

https://go-firebird.fly.dev/api/testing/geocoding


> Going to wipe the db 

Now we can bring it all together with the database and see it be saved live
https://go-firebird.fly.dev/api/firebird/bluesky



I also made a listner on the testing route: 
http://localhost:3000/testing
