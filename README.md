README
------

Please maintain TODO/README

INSTALLATION
------------

go get github.com/seedboxtech/gojira


USAGE
-----

gojira -h


gojira -u username -p password -n --stdin 
-n: don't check ssl validity of certificate. 
 
--stdin: grab a json file from stdin. 
Pipe a modified query like this to gojira to get the issues for the current sprint. 
curl -k -u UNAME:PWORD "https://jira.yourserver.com/rest/api/2/search?jql=sprint+in+openSprints()+AND+project+=+'Traffic+Division'" 

-List tasks 
gojira list     //ordered by ranking by default 
gojira list -c  //current sprint 
gojira list -o  //open tasks 
gojira list -p "Project Name" //Project filter 
gojira list -i "Issue-ID" // Filter by issue-id 

Mix and match previous three options for fun and profit. 


Get the worklog for a story 
./gojira -n -c -u otremblay -p REDACTED | grep 1789 | grep -v "Story" | while read l; do x=$(echo $l | grep -o 'TDIV-[0-9]\+'); curl -k -u otremblay:REDACTED "https://jira.yourserver.com/rest/api/2/issue/$x/worklog"; done 

CONFIGURATION
-------------
gojira supports a configuration file ~/.gojirarc, see the example file provided: example.gojirarc 

LICENSE
-------

BSD-3


TIPS
----

If you find this software useful in any way and wish to reward the work that went in it, I accept donations in the form of Dogecoins at this address: DQpQKEVThLMMfV8HWg69wjnAPgzNLJyNem 
