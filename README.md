# yodahunters

**TODO:**

- Add `SetLevel` function to internal/log

# Database Notes

Planning to go with more tables than fewer. Going to have
a table for threads, icons, ratings and users to start. Each thread will
have its own table to hold the posts.

**Threads**
ID - int
Title - TEXT
Desc - TEXT
Author ID - int
Icon Link - TEXT
created_at - TIMESTAMPTZ
Replies - int

**Users**
ID - int
username - varchar(20)
password_hash - varchar(100)
email - text
reg_date - timestamptz
profile_pic - text

**Thread Table (Posts)**
ID - int
content - text
reply ID - int // the id of the post this post is replying to, if there is one
author - int // author's user id
timestamp - timestamptz

(Future Work)
**Ratings**

**Icons**
