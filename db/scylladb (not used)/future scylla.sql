/* 
Why Scylla DB:

Super fast
10 X faster than Cassandra and no JVM, all C++
Used by Apple (100,000 servers!), Google Search, Uber, Spotify and many others as main storage
Proven masterless automatic horizontal scalability that is also data-centre aware 
Easy CQL query language
Time Series data using TimeUUID
*/

create keyspace if not exists ur with replication = {
  'class': 'SimpleStrategy',
  'replication_factor': '1'
};

use ur;

create type if not exists fullname (
  first_name text,
  middle_name text,
  last_name text
);

create type if not exists address (
    street1 text ,
    street2 text ,
    suburb text ,
    state text ,
    postcode text , 
    country text ,
);

-- drop table users;

create table if not exists users (
    user_id timeuuid,
    full_name frozen <fullname> static,
    email text  static, 

    ur_roles set<text> static, -- UPDATE users SET ur_roles = ur_roles - {'seller'}  where user_id = 62c36092-82a1-3a00-93d1-46196ee77204
                               -- CREATE INDEX users_ur_roles ON users(ur_roles) -- SELECT * FROM users WHERE ur_roles CONTAINS 'admin'
    date_of_birth timestamp static, 
    facebook_id text static,
    paypal_id text static,
    credit_card text static, -- encrypted JSON
    current_feedback_rating int static, --e.g. 4 for 4 stars
    current_reputation_percentage int static, --e.g. 97 for 'extremely reliable'

    addresses map<text, frozen <address>> static, --e.g. <"home", address>
    address_default frozen<address> static,
    phones set<text> static,
    last_login map<text,
      frozen<tuple<timestamp,inet>>
    > static, -- UPDATE users USING TTL 31536000 SET last_login['iPhone7'] = {'2015-07-01 11:17:42', '192.168.22.3'} WHERE user_id = 62c36092-82a1-3a00-93d1-46196ee77204;

    profile_id int,
    profile_name text,

    PRIMARY KEY (user_id, profile_id)
);






create table feedback (
  bucket_id text, -- auto partition creation by month e.g. '2016-12' for queries like 
  --                 ... top 30 days where bucket_id in ('2016-12', '2015-12'). 
  --                 TODO: Write Go function that generates the 'in' statement for a period of exactly 12 months before now
  --                 so that the correct partitions are searched for rating calculations
  id timeuuid,
  user_to timeuuid,
  user_from timeuuid,
  value int
  comment text
  primary key ((bucket_id, id, user_to), user_from) 
) with clustering order by (id desc);

--we'll need to search feedback by user_from so create secondary index:
--create index feedback_user_from ON feedback(user_from)


insert into users (user_id, first_name, last_name, date_of_birth, email, 
                  street1, street2, suburb, state, postcode, country, 
                  facebook_id, paypal_id, credit_card)

values (62c36092-82a1-3a00-93d1-46196ee77204, {first_name: 'Nuri', middle_name: NULL, last_name: 'Kevenoglu'}, 
          'nuri@universalreputations.com', 
          {'root','admin','ambassador','buyer','seller'}
          '1963-12-18',
          '19 Streamdale Grove', NULL, 'Warriewood', 'NSW', '2102', 'Australia', 
          NULL, NULL, NULL) -- these 3 columns are JSON strings
if not exists


insert into users (user_id, role_id, role_name) values (62c36092-82a1-3a00-93d1-46196ee77204,1,'Admin')
if not exists

insert into users (user_id, profile_id, profile_name) values (62c36092-82a1-3a00-93d1-46196ee77204,1,'Singing Teacher')
if not exists
