Woofie
======
This is a webserver that plays sounds when triggered from an HTTP client. The
intent is to integrate it into an IoT solution with a remote motion sensor to
play dog barks to (hopefully) scare off crooks.

It uses the GoFlaCook cookbook code to help play FLAC files and has a decently
sophisticated bit of business logic to figure out when to play.


Installation
------------
The easiest way is to just:

`go get github.com/wjblack/woofie/...`
`go install github.com/wjblack/woofie/...`

For sure you'll need the portaudio dev files (e.g. apt-get install
portaudio19-dev or similar).

If you want to test this (especially on a different arch), I definitely
recommend running "go test -v github.com/wjblack/woofie".


Usage
-----
You could just start bin/woofie with no params, or:

`bin/woofie --help`
`bin/woofie --woofdir=/path/to/FLAC/dir/`

...the latter of which will start the server on port 40080 with a default path.

The basics like TCP port, what the assumed path is (handy if e.g. you're behind
a load balancer that passes paths through), and so on are in there and fairly
straightforward.  The business logic, however, is a bit more complex and is
explained below.

The schedule is decently sophisticated and is intended to be a sort of "during
business hours" disablement of the virtual dog.  For example, if you want the
dog to shut up totally between 9 and 5 on weekdays and 12 to 3:30 on Saturdays:

`bin/woofie --schedule='1-5=9-17,6=12-15:30'`

The idea is that the day of the week range is kinda like UNIX cron, where zero
is Sunday, one is Monday and so on.  Ranges like '1-5' work, and the hour
range can be single hours or full HH:MM format (mix and match).  Plus you can
overlap or have holes no problem, so this is legit:

`bin/woofie --schedule='1-5=9-11:30,1-5=12:30-17,4=10-18'`

...which would be:

* Monday, Tuesday, Wednesday, Friday 9-11:30 and 12:30-5 and
* Thursday 9-6

Note all times are 24-hour time.


Business Logic
--------------
We're trying to simulate how a dog thinks here and to try not to be too
annoying if the dog's being constantly stimulated.  By default, the algorithm:

* Barks if motion is detected (i.e. the HTTP /.../on request is called).
* Doesn't bark if it's been barking a lot lately (e.g. for the last 5 minutes
  straight).
* Might decide to randomly bark anyway (fickle, these virtual doggies), and:
* Therefore has a scoring system to see how annoying it's been lately.

The algorithm can be adjusted with a few parameters:

`--resolution=15`

Resolution is how long to bark at minimum.  Basically, we'll repeatedly play
the FLAC files until at least this much time has passed.

`--horizon=30`

Horizon tells the algorithm how many minutes to look in the past for.  The
shorter the horizon, the quicker the annoyance threshold drops off.

`--score=150`

The score gives a max point after which we won't bark again for a while.  The
score is calculated by assigning a point value to each bark entry in the logs
in decaying proportion based on how long ago the bark was (this minute = max
points, near the horizon = min points).  So if we barked a bunch in the
last half hour, we could decide to cool it for a bit (or if we have less more
recently).

`--randomfactor=5`

This is a percentage chance that the virtual dog will decide to forego the
above algorithm and bark anyway if motion is detected.  That way the dog
lets his presence known if e.g. there's frequent motion detected) without being
too annoying.

In general, the default settings should be decently doglike without being
annoying--the dog will get "tired" after 5 minutes of barking for 15-30 minutes
and will decide to bark anyway 5% of the time.


Future Plans
------------
Since this server was designed to be used in conjunction with a client, I'll
have a NodeMCU implementation of a client that polls GPIO soon...
