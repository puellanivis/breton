# Why?

The internal golang "os/user" library will try a domain lookup, to fill in some user information, even if cached information is available.

Depending on response rates, this can add minutes to execution time at start up, because glog looks up the user name, which with the default library, fills in all user information, even though all it needs is the username.
