Croncape
========

Croncape wraps commands run as cron jobs to send emails **only** when an error
or a timeout has occurred.

Out of the box, crontab can send an email when a job generates output. But a
command is not necessarily unsuccessful "just" because it used the standard or
error output. Checking the exit code would be better, but that's not how
crontab was [standardized][1].

Croncape takes a different approach by wrapping your commands to only send an
email when the command returns a non-zero exit code.

Croncape plays well with crontab as it never outputs anything except when an
issue occurs in Croncape itself (like a misconfiguration for instance), in
which case crontab would send you an email.

Installation
------------

Download the [binaries][4] or `go get github.com/sensiocloud/croncape`.

Usage
-----

When adding a command in crontab, prefix it with `croncape`:

    0 6 * * * croncape -e "sysadmins@example.com" ls -lsa

That's it!

You can also send emails to more than one user by separating emails with a comma:

    0 6 * * * croncape -e "sysadmins@example.com,sys@foo.org" ls -lsa

Besides sending emails, croncape can also kill the run command after a given
timeout, via the `-t` flag (disabled by default):

    0 6 * * * croncape -e "sysadmins@example.com" -t 2h ls -lsa

If you want to send emails even when commands are successful, use the `-v` flag
(useful for testing).

Use the `-h` flag to display the full help message.

Croncape is very similar to [cronwrap][2], with some differences:

 * No dependencies (cronwrap is written in Python);

 * Kills a command on a timeout (cronwrap just reports that the command took
   more time to execute);

 * Tries to use `sendmail` or `mail` depending on availability (cronwrap only
   works with `sendmail`).

For a simpler alternative, have a look at [cronic][3].

[1]: http://pubs.opengroup.org/onlinepubs/9699919799/utilities/crontab.html
[2]: https://pypi.python.org/pypi/cronwrap/1.4
[3]: http://habilis.net/cronic/
[4]: https://github.com/sensiocloud/croncape/releases
