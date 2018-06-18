Croncape
========

Croncape wraps commands run as cron jobs to send emails **only** when an error
or a timeout has occurred.

Out of the box, crontab can send an email when a job [generates output][5]. But
a command is not necessarily unsuccessful "just" because it used the standard
or error output. Checking the exit code would be better, but that's not how
crontab was [standardized][1].

Croncape takes a different approach by wrapping your commands to only send an
email when the command returns a non-zero exit code.

Croncape plays well with crontab as it never outputs anything except when an
issue occurs in Croncape itself (like a misconfiguration for instance), in
which case crontab would send you an email.

Installation
------------

Download the [binaries][4] or `go get github.com/symfonycorp/croncape`.

Usage
-----

When adding a command in crontab, prefix it with `croncape`:

    MAILTO=sysadmins@example.com
    0 6 * * * croncape ls -lsa

That's it!

Note that the `MAILTO` environment variable can also be defined globally in
`/etc/crontab`; it supports multiple recipients by separating them with a comma.

If you need to use "special" shell characters in your command (like `;` or `|`),
don't forget to quote it:

    0 6 * * * croncape "ls -lsa | true"

Besides sending emails, croncape can also kill the run command after a given
timeout, via the `-t` flag (disabled by default):

    0 6 * * * croncape -t 2h ls -lsa

If you want to send emails even when commands are successful, use the `-v` flag
(useful for testing).

Use the `-h` flag to display the full help message.

Croncape is very similar to [cronwrap][2], with some differences:

 * No dependencies (cronwrap is written in Python);

 * Kills a command on a timeout (cronwrap just reports that the command took
   more time to execute);

 * Tries to use `sendmail` or `mail` depending on availability (cronwrap only
   works with `sendmail`);

 * Reads the email from the standard crontab `MAILTO` environment variable
   instead of a `-e` flag.

For a simpler alternative, have a look at [cronic][3].

[1]: http://pubs.opengroup.org/onlinepubs/9699919799/utilities/crontab.html
[2]: https://pypi.python.org/pypi/cronwrap/1.4
[3]: http://habilis.net/cronic/
[4]: https://github.com/symfonycorp/croncape/releases
[5]: https://xkcd.com/1728/
