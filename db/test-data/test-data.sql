/*
This script is not yet complete but currently can be used to test.
It creates sample users with profiles.

Name	            username	            password	credit card
Alex Merphy	        amerphy2118@gmail.com	UR-am2118	5206770789742227
Bernard Smith	    bernards2118@gmail.com	UR-bs2118	5459005252006910
Emily Perkins	    emilyp2118@gmail.com	UR-ep2118	5436749282007860
Eugene Newman	    eugenen2118@gmail.com	UR-en2118	5575208276646370
Julian Dickons	    juliand2118@gmail.com	UR-jd2118	4532336371090678
Ron Oswald	        roswald2118@gmail.com	UR-ro2118	4024007176812690
Russell Myers	    russellm2118@gmail.com	UR-rm2118	4556073945242847
Samantha Roberts	samanthar2118@gmail.com	UR-sr2118	4024007198541749
Shirley Vincent	    shirleyv2118@gmail.com	UR-sv2118	4621983152939482

Please delete the folder /static/images/profile-photos and then copy 
the folder /db/test-data/profile-photos (incluing all subfolders) to become /static/images/profile-photos.

*/

TRUNCATE TABLE serviceinputtype CASCADE;
INSERT INTO serviceinputtype (service_input_id, from_date, to_date, from_time, to_time, frequency_unit, total_price) VALUES 
(0, false, false, false, false, false, true),
(1, true, false, true, true, false, false),
(2, true, true, true, true, false, false),
(3, true, false, true, true, true, false);

-- Setup sequence so it doesn't break next entry. Should be equal to last used id plus 1
ALTER SEQUENCE serviceinputtype_service_input_id_seq RESTART WITH 4;


/****** Insert servicecategory data ******/
-------------------------------------------
TRUNCATE TABLE servicecategory CASCADE;
INSERT INTO servicecategory (service_id, service_name, service_style_key, parent_id) VALUES 
    (0, 'N/A', null, 0), --  This row is used for all profile types except service provider
    (1, 'Care/Help', null, 0),
        (2, 'Babysitter', 1, 1),
        (3, 'Nanny', 3, 1),
        (4, 'Caretaker', 3, 1),
        (5, 'Pet Care', 2, 1),
    (6, 'Drive/Deliver', null, 0),
        (7, 'Delivery man', 1, 6),
        (8, 'Driver', 2, 6),
        (9, 'Removalist', 2, 6),        
    (10, 'Maintain/Repair', null, 0),
        (11, 'Electrician', 1, 10),
        (12, 'Plumber', 1, 10),    
        (13, 'Gardener', 3, 10),
        (14, 'Handy man', 1, 10),        
        (15, 'Lawn mower', 3, 10),
        (16, 'Computers', 1, 10),
    (17, 'Cook/Clean', null, 0),
        (18, 'Home cleaner', 1, 17),        
        (19, 'Office cleaner', 1, 17), 
        (20, 'Cook', 1, 17),
        (21, 'Laundry/Ironing ', 1, 17),
    (22, 'Health/Sports', null, 0),
      (23, 'Personal Trainer', 1, 22),
      (24, 'Group Trainer', 1, 22),
      (25, 'Sport Coach', 1, 22),
      (26, 'Gym Rental', 1, 22),
      (27, 'Weight loss', 2, 22),
    (28, 'Learning', null, 0),
      (29, 'Music tutor', 1, 28),
      (30, 'School tutor', 1, 28),
      (31, 'Language tutor', 1, 28),
    (32, 'Other', null, 0),
      (33, 'Other', 1, 32);

-- Setup sequence so it doesn't break next entry. Should be equal to last used id plus 1
ALTER SEQUENCE servicecategory_service_id_seq RESTART WITH 33;    


/****** Insert sda_types data ******/
-------------------------------------------
TRUNCATE TABLE sda_types CASCADE;
INSERT INTO sda_types (sda_id,ref_id, profile_type,sda_list) VALUES 
(1,2,'p', '{"RELIABLE", "FRIENDLY", "TRUSTWORTHY", "GOOD WITH KIDS","PRESENTABLE", "LIKEABLE", "PUNCTUAL", "CLEAN", "GOOD COMMUNICATION", "INITIATIVE"}'),
(2,3, 'p', '{"RELIABLE", "FRIENDLY", "TRUSTWORTHY", "GOOD WITH CAR","PRESENTABLE", "LIKEABLE", "PUNCTUAL", "CLEAN", "GOOD COMMUNICATION", "INITIATIVE"}'),
(3,5,'p', '{"RELIABLE", "FRIENDLY", "TRUSTWORTHY", "GOOD WITH KIDS","PRESENTABLE", "LIKEABLE", "PUNCTUAL", "CLEAN", "GOOD COMMUNICATION", "INITIATIVE"}'),
(4,7,'p', '{"RELIABLE", "FRIENDLY", "TRUSTWORTHY", "DILIGENT","PRESENTABLE", "LIKEABLE", "PUNCTUAL", "CLEAN", "GOOD COMMUNICATION", "INITIATIVE"}'),
(5,0,'b', '{"PUNCTUAL PAYMENT", "UNDERSTANDING", "FRIENDLY", "COURTEOUS","CLEAR EXPECTATIONS", "GENEROUS", "CONSIDERATE", "POLITE", "TOLERANT", "ORGANISED"}'),
(6,0, 's', '{"PUNCTUAL PAYMENT", "UNDERSTANDING", "FRIENDLY", "COURTEOUS","CLEAR EXPECTATIONS", "GENEROUS", "CONSIDERATE", "POLITE", "TOLERANT", "ORGANISED"}'),
(7,8,'p', '{"RELIABLE", "FRIENDLY", "TRUSTWORTHY", "GOOD WITH KIDS","PRESENTABLE", "LIKEABLE", "PUNCTUAL", "CLEAN", "GOOD COMMUNICATION", "INITIATIVE"}'),
(8,9,'p', '{"RELIABLE", "FRIENDLY", "TRUSTWORTHY", "GOOD WITH KIDS","PRESENTABLE", "LIKEABLE", "PUNCTUAL", "CLEAN", "GOOD COMMUNICATION", "INITIATIVE"}'),
(9,11,'p', '{"RELIABLE", "FRIENDLY", "TRUSTWORTHY", "GOOD WITH KIDS","PRESENTABLE", "LIKEABLE", "PUNCTUAL", "CLEAN", "GOOD COMMUNICATION", "INITIATIVE"}'),
(10,12,'p', '{"RELIABLE", "FRIENDLY", "TRUSTWORTHY", "GOOD WITH KIDS","PRESENTABLE", "LIKEABLE", "PUNCTUAL", "CLEAN", "GOOD COMMUNICATION", "INITIATIVE"}'),
(11,13,'p', '{"RELIABLE", "FRIENDLY", "TRUSTWORTHY", "GOOD WITH KIDS","PRESENTABLE", "LIKEABLE", "PUNCTUAL", "CLEAN", "GOOD COMMUNICATION", "INITIATIVE"}'),
(12,14,'p', '{"RELIABLE", "FRIENDLY", "TRUSTWORTHY", "GOOD WITH KIDS","PRESENTABLE", "LIKEABLE", "PUNCTUAL", "CLEAN", "GOOD COMMUNICATION", "INITIATIVE"}'),
(13,15,'p', '{"RELIABLE", "FRIENDLY", "TRUSTWORTHY", "GOOD WITH KIDS","PRESENTABLE", "LIKEABLE", "PUNCTUAL", "CLEAN", "GOOD COMMUNICATION", "INITIATIVE"}'),
(14,16,'p', '{"RELIABLE", "FRIENDLY", "TRUSTWORTHY", "GOOD WITH KIDS","PRESENTABLE", "LIKEABLE", "PUNCTUAL", "CLEAN", "GOOD COMMUNICATION", "INITIATIVE"}'),
(15,18,'p', '{"RELIABLE", "FRIENDLY", "TRUSTWORTHY", "GOOD WITH KIDS","PRESENTABLE", "LIKEABLE", "PUNCTUAL", "CLEAN", "GOOD COMMUNICATION", "INITIATIVE"}'),
(16,19,'p', '{"RELIABLE", "FRIENDLY", "TRUSTWORTHY", "GOOD WITH KIDS","PRESENTABLE", "LIKEABLE", "PUNCTUAL", "CLEAN", "GOOD COMMUNICATION", "INITIATIVE"}'),
(17,20,'p', '{"RELIABLE", "FRIENDLY", "TRUSTWORTHY", "GOOD WITH KIDS","PRESENTABLE", "LIKEABLE", "PUNCTUAL", "CLEAN", "GOOD COMMUNICATION", "INITIATIVE"}'),
(18,21,'p', '{"RELIABLE", "FRIENDLY", "TRUSTWORTHY", "GOOD WITH KIDS","PRESENTABLE", "LIKEABLE", "PUNCTUAL", "CLEAN", "GOOD COMMUNICATION", "INITIATIVE"}'),
(19,23,'p', '{"RELIABLE", "FRIENDLY", "TRUSTWORTHY", "GOOD WITH KIDS","PRESENTABLE", "LIKEABLE", "PUNCTUAL", "CLEAN", "GOOD COMMUNICATION", "INITIATIVE"}'),
(20,24,'p', '{"RELIABLE", "FRIENDLY", "TRUSTWORTHY", "GOOD WITH KIDS","PRESENTABLE", "LIKEABLE", "PUNCTUAL", "CLEAN", "GOOD COMMUNICATION", "INITIATIVE"}'),
(21,25,'p', '{"RELIABLE", "FRIENDLY", "TRUSTWORTHY", "GOOD WITH KIDS","PRESENTABLE", "LIKEABLE", "PUNCTUAL", "CLEAN", "GOOD COMMUNICATION", "INITIATIVE"}'),
(22,26,'p', '{"RELIABLE", "FRIENDLY", "TRUSTWORTHY", "GOOD WITH KIDS","PRESENTABLE", "LIKEABLE", "PUNCTUAL", "CLEAN", "GOOD COMMUNICATION", "INITIATIVE"}'),
(23,27,'p', '{"RELIABLE", "FRIENDLY", "TRUSTWORTHY", "GOOD WITH KIDS","PRESENTABLE", "LIKEABLE", "PUNCTUAL", "CLEAN", "GOOD COMMUNICATION", "INITIATIVE"}'),
(24,29,'p', '{"RELIABLE", "FRIENDLY", "TRUSTWORTHY", "GOOD WITH KIDS","PRESENTABLE", "LIKEABLE", "PUNCTUAL", "CLEAN", "GOOD COMMUNICATION", "INITIATIVE"}'),
(25,30,'p', '{"RELIABLE", "FRIENDLY", "TRUSTWORTHY", "GOOD WITH KIDS","PRESENTABLE", "LIKEABLE", "PUNCTUAL", "CLEAN", "GOOD COMMUNICATION", "INITIATIVE"}'),
(26,31,'p', '{"RELIABLE", "FRIENDLY", "TRUSTWORTHY", "GOOD WITH KIDS","PRESENTABLE", "LIKEABLE", "PUNCTUAL", "CLEAN", "GOOD COMMUNICATION", "INITIATIVE"}'),
(27,33,'p', '{"RELIABLE", "FRIENDLY", "TRUSTWORTHY", "GOOD WITH KIDS","PRESENTABLE", "LIKEABLE", "PUNCTUAL", "CLEAN", "GOOD COMMUNICATION", "INITIATIVE"}');
-- Setup sequence so it doesn't break next entry. Should be equal to last used id plus 1
ALTER SEQUENCE sda_types_sda_id_seq RESTART WITH 28;    



/****** Insert User data ******/
---------------------------------
-- Please copy all folders in /app/db/profile-photos/* to /static/images/profile-photos/*

TRUNCATE TABLE public.users CASCADE;
INSERT INTO public.users (user_id, first_name, middle_name, last_name, email, date_of_birth, facebook_id, paypal_id, credit_card_id, password, phones, last_login, created, updated, photo_url, roles, postcode, location, credit_card_mask) VALUES 

--FAKE MASTERCARD 5206770789742227 from http://www.getcreditcardnumbers.com/
(1, 'Alex', '', 'Merphy', 'amerphy2118@gmail.com', NULL, '', NULL, 'CARD-5PN74329B3831222GLDPZKWY', '$2a$10$Kh5o5yuD2wjGSfHvq/Ho3u9fCL5vGr0ZNe0zubhKHj1PP4H9by/I2', '', NULL, '2015-03-11 11:04:38.745', '2017-04-01 11:53:42.166', '4798b947-e51f-4ea9-8c2c-4c5fa06658f8-user-alex-merphy-270x270.jpg', '{user}', '2102', '0101000020E610000092B245D26EE96240B404190115D840C0', '52xx xxxx xxxx 2227'),
--FAKE MASTERCARD 5459005252006910
(2, 'Bernard', '', 'Smith', 'bernards2118@gmail.com', NULL, '', NULL, 'CARD-02671689BV576963YLDQH3NI', '$2a$10$qZ522gH/OhDCsxHLIGIoaOmtbPu.kmLepho.Uv3hz5YkBn7vxeC1u', '', NULL, '2015-08-22 04:23:36.189', '2017-04-02 04:33:19.776', '8a1eefd0-6e55-46e7-945f-18f27af9ad39-user-bernard-smith-270x270.jpg', '{user}', '2100', '0101000020E6100000B98C9B1A68E86240234BE658DEE140C0', '54xx xxxx xxxx 6915'),
--FAKE MASTERCARD 5436749282007868
(3, 'Emily', '', 'Perkins', 'emilyp2118@gmail.com', NULL, '', NULL, 'CARD-8PB06059BL369800PLDQIFTY', '$2a$10$UyKJHdoVxfQu/VgqDpnVLeb7IvhhPkQCzFaKmBeoOcg8qIBQFAWLC', '', NULL, '2016-07-13 04:44:53.203', '2017-04-02 04:46:50.527', 'fbb4b405-92fc-4eed-93f1-2cb03d8e5d8a-user-emily-perkins-270x270.jpg', '{user}', '2098', '0101000020E6100000CD3D247CEFCF2240664D2CF015994740', '54xx xxxx xxxx 7868'),
--FAKE MASTERCARD 5575208276646370
(4, 'Eugene', '', 'Newman', 'eugenen2118@gmail.com', NULL, '', NULL, 'CARD-4CA8041360051235GLDQIHUY', '$2a$10$N52TlzWET7NNw3NdCw/nt.u8l6MC/HE4jegoK5PCB5NBJm3yzTi5G', '', NULL, '2017-04-02 04:49:37.050', '2017-04-02 04:51:10.787', 'e3e25a88-4d72-4caa-8a76-927d79aa52b1-user-eugene-newman-270x270.jpg', '{user}', '2099', '0101000020E61000008DD47B2A27E9624083DE1B4300E040C0', '55xx xxxx xxxx 6372'),
--FAKE VISA 4532336371090678
(5, 'Julian', '', 'Dickons', 'juliand2118@gmail.com', NULL, '', NULL, 'CARD-21F35764GP165771DLDQIMPA', '$2a$10$6kks5XnfRutNRV/Q0g06/e8JeWQhKAPGm28QE0jIzKyjlmg92tA0C', '', NULL, '2016-04-12 04:59:38.468', '2017-04-02 05:01:28.212', '4e9de917-837f-44c1-aa60-40d6a043ad6a-user-julian-dickons-270x270.jpg', '{user}', '2099', '0101000020E61000008DD47B2A27E9624083DE1B4300E040C0', '45xx xxxx xxxx 0678'),
--FAKE VISA 4024007176812690
(6, 'Ron', '', 'Oswald', 'roswald2118@gmail.com', NULL, '', NULL, 'CARD-6K495083K50002358LDQIOCA', '$2a$10$NTaoChmxcNaJ4ej0kj.oIOUgEr70cGj46K/CDaTBvnHqI.bnH43pS', '', NULL, '2017-01-07 05:03:06.770', '2017-04-02 05:04:51.429', '206328c5-cceb-4d84-aa25-428ab6e6f984-user-ron-oswald-270x270.jpg', '{user}', '2089', '0101000020E6100000F96706F101E76240A227655243EB40C0', '40xx xxxx xxxx 2690'),
--FAKE VISA 4556073945242847
(7, 'Russell', '', 'Myers', 'russellm2118@gmail.com', NULL, '', NULL, 'CARD-67308398AC681670FLDQIPXA', '$2a$10$Q6kREXbB04IzoDICozyUkOxAL2uKvUQccr.D36jGMp7SRW6Y5ww4m', '', NULL, '2014-02-02 05:06:49.861', '2017-04-02 05:08:23.637', 'c71eb6f7-1ff9-45f7-ac25-60ff3976ebf5-user-russell-myers-270x270.jpg', '{user}', '2088', '0101000020E6100000C1E10511A9E76240B1A888D349EA40C0', '45xx xxxx xxxx 2847'),
--FAKE VISA 4024007198541749
(8, 'Samantha', '', 'Roberts', 'samanthar2118@gmail.com', NULL, '', NULL, 'CARD-99P442380A586282BLDQIRKI', '$2a$10$pjkJzrXKCDnQwU5rxlPJ/einBIEshJfH894yATzCpMbhZoBuBv3JO', '', NULL, '2015-08-23 05:10:03.222', '2017-04-02 05:11:48.737', 'bb548aba-b96c-4ec7-87e6-b5f9ebcce282-user-samantha-roberts-270x270.jpg', '{user}', '2100', '0101000020E6100000B98C9B1A68E86240234BE658DEE140C0', '40xx xxxx xxxx 1749'),
--FAKE VISA 4621983152939482
(9, 'Shirley', '', 'Vincent', 'shirleyv2118@gmail.com', NULL, '', NULL, 'CARD-02E42987BB932490FLDQITLI', '$2a$10$NuOxCMnXA9Y8maj7gIs.ouibT9ohvaGweu2ycs9oOBw3A/MMNtwqi', '', NULL, '2017-12-18 05:14:29.175', '2017-04-02 05:16:09.008', '992d45cb-1331-4f12-94cd-dec2e75b5384-user-shirley-vincent-270x270.jpg', '{user}', '2087', '0101000020E61000008E91EC11EAE66240392BA226FAE240C0', '46xx xxxx xxxx 9482'),
--No  Credit card
(10, 'Ivan', '', 'Cheng', 'blueskymobile@hotmail.com', NULL, '', NULL, NULL, '$2a$10$1dki8zKFNAQChwlc/WuL7O52Q6TH1WTCsE.VI3NpGMNRgb36Dx/F2', '', NULL, '2017-04-26 06:07:49.247', '2017-04-26 06:07:49.248', '992d45cb-1331-4f12-94cd-dec2e75b5384-user-shirley-vincent-270x270.jpg', '{user}', '2850', '0101000020E61000002EC6C03A0EB26240236937FA984740C0', ''),
--No  Credit card
(11, 'Liu', '', 'Jin',  'liu0514jin@gmail.com',  NULL, '', NULL, '', '$2a$10$dVldBTr.L4qp/7HKrxgGue7EnB2SJkYgZ0skFE5NUHFAGZmufFdQO', '', NULL, '2017-04-26 06:57:00.623723', '2017-04-26 06:57:15.112945', '', '{user}', '2102', '0101000020E610000092B245D26EE96240B404190115D840C0', ''),
--No  Credit card, Facebook only
(12, 'Nuri', '', 'Kevenoğlu', 'smyrnian@hotmail.com', NULL, '10154493605467490', NULL, '', '', '', NULL, '0001-01-01 00:00:00.000', '2017-04-27 05:00:00.949', '10154493605467490-Picture.jpg', '{user}', '2087', '0101000020E61000008E91EC11EAE66240392BA226FAE240C0', '');


-- Setup sequence so it doesn't break next entry. Should be equal to last used id plus 1
ALTER SEQUENCE users_user_id_seq RESTART WITH 13;



/****** Insert profile data ******/
---------------------------------
TRUNCATE table public.profiles cascade;
/* profile_type values:
    'w' : Wanted is a customer advertising a job/help they need (like Airtasker) or something they want to buy (send one-off or reoccurring payment(s))
    'r' : Rents or hires things (receive one-off or reoccurring payment(s))
    's' : Sells things (receive one off payment)
    'b' : Buys things (send one-off payment). This is the default customer profile created automatically for each user.
    'p' : Provider of a service (receive one-off or reccurring payment(s))
*/

--'Alex Merphy'
INSERT INTO public.profiles
(profile_id, user_id, photo_url, description, feedback_rating, reputation_status, fee, created, updated, payment_notes, title, profile_type, heading, service_category, profile_uuid) VALUES
(1, 1, '4798b947-e51f-4ea9-8c2c-4c5fa06658f8-user-alex-merphy-270x270.jpg', 'I love computers and everything to do with them.', 0, 0, '', '2017-04-02 06:02:29.944', '2017-04-05 11:11:25.573', '', '', 'b', 'IT Geek', 0, '7c2419a2-2b35-11e7-bde0-f40f24220798'),
(2, 1, 'a766a5df-d718-4253-8f7f-7e0b7e23f414-alex-merphy.jpg', 'I fix all kinds of computers; PCs, Macs, Macbooks, Server issues, you name it. Very experienced and highly regarded by clients.', 95, 96, '$100/hr', '2017-04-05 10:59:01.292', '2017-04-05 11:04:00.110', '$50 off if I can''t fix it on the spot!', 'I fix computers', 'p', 'I can fix any computer!', 16, '7c241f10-2b35-11e7-bde0-f40f24220798'),
(3, 1, '2a18080e-0449-4715-9e52-d3bc9d86afac-computer-in-a-shopping-cart.jpg', 'I sell new and refurbished PCs, Macs, Windows notebooks, Macbooks and related accessories.', 90, 80, 'From $200', '2017-04-05 11:07:48.915', '2017-04-05 11:08:20.669', '', 'New and used computers', 's', 'Best value computers in town!', 0, '7c241fb0-2b35-11e7-bde0-f40f24220798'),

--'Bernard Smith'
(4, 2, '8a1eefd0-6e55-46e7-945f-18f27af9ad39-user-bernard-smith-270x270.jpg', 'I run a plumbing/electrical agency. I''m always interested in ways to improve the business.', 0, 0, '', '2017-04-02 06:02:29.944', '2017-04-05 11:11:25.573', '', '', 'b', 'Honest perfectionist...', 0,'7c242000-2b35-11e7-bde0-f40f24220798'),
(5, 2, 'a1294a30-55cb-44c5-a003-2946bd734430-plumber.jpg', 'Licensed local plumber with 20 years experience.', 75, 88, '$80 +GST', '2015-04-13 04:28:45.124', '2017-04-13 05:19:58.437', '', 'Your local plumber', 'p', 'Cheap and efficient', 12,'7c242046-2b35-11e7-bde0-f40f24220798'),
(6, 2, 'a75971b4-5a27-4d1c-80d6-78a8e1655a71-electrician.jpg', 'Local sparkie at your service', 65, 65, 'From $75', '2016-08-13 05:23:19.209', '2017-04-13 05:51:00.726', '', 'Electrician', 'p', 'Best sparkie in your area!', 11,'7c24208c-2b35-11e7-bde0-f40f24220798'),

--'Emily Perkins'
(7, 3, 'fbb4b405-92fc-4eed-93f1-2cb03d8e5d8a-user-emily-perkins-270x270.jpg', 'I love working and playing with children', 0, 0, '', '2016-12-02 16:02:29.944', '2017-04-05 11:11:25.573', '', '', 'b', 'Those wide open eyes when listening to a story...', 0,'7c2420d2-2b35-11e7-bde0-f40f24220798'),
(8, 3, 'e28b8279-0cca-41cf-ac2d-69de095acf15-profile-emily-perkins-270x270.jpg', 'I am a level 3C qualified preschool tutor with 5 years experience in the Northern Beaches. References available on request.', 0, 0, '$30/hr', '2017-04-13 09:44:25.491', '2017-04-13 09:44:25.491', '', 'Preschool Tutor', 'p', 'Give your child a head start', 30,'7c242118-2b35-11e7-bde0-f40f24220798'),

--'Eugene Newman'
(9, 4, 'e3e25a88-4d72-4caa-8a76-927d79aa52b1-user-eugene-newman-270x270.jpg', 'I love driving', 0, 0, '', '2016-12-02 16:02:29.944', '2017-04-05 11:11:25.573', '', '', 'b', '', 0,'7c24215e-2b35-11e7-bde0-f40f24220798'),

--'Julian Dickons'
(10, 5, '4e9de917-837f-44c1-aa60-40d6a043ad6a-user-julian-dickons-270x270.jpg', 'I am a Handy man', 0, 0, '', '2016-12-02 16:02:29.944', '2017-04-05 11:11:25.573', '', '', 'b', '', 0,'7c2421a4-2b35-11e7-bde0-f40f24220798'),

--'Ron Oswald'
(11, 6, '206328c5-cceb-4d84-aa25-428ab6e6f984-user-ron-oswald-270x270.jpg', 'I am a Office cleaner', 0, 0, '', '2016-12-02 16:02:29.944', '2017-04-05 11:11:25.573', '', '', 'b', '', 0,'7c2421ea-2b35-11e7-bde0-f40f24220798'),

--'Russell Myers'
(12, 7, 'c71eb6f7-1ff9-45f7-ac25-60ff3976ebf5-user-russell-myers-270x270.jpg', 'Personal Trainer', 0, 0, '', '2016-12-02 16:02:29.944', '2017-04-05 11:11:25.573', '', '', 'b', '', 0,'7c242226-2b35-11e7-bde0-f40f24220798'),

--'Samantha Roberts'
(13, 8, 'bb548aba-b96c-4ec7-87e6-b5f9ebcce282-user-samantha-roberts-270x270.jpg', 'I love babies and playing with children', 0, 0, '', '2016-12-02 16:02:29.944', '2017-04-05 11:11:25.573', '', '', 'b', '', 0,'7c24226c-2b35-11e7-bde0-f40f24220798'),

--'Shirley Vincent'
(14, 9, '992d45cb-1331-4f12-94cd-dec2e75b5384-user-shirley-vincent-270x270.jpg', '', 45, 78, '', '2016-12-02 16:02:29.944', '2017-04-05 11:11:25.573', '', '', 'b', '', 0,'7c2422d0-2b35-11e7-bde0-f40f24220798'),

--'Ivan Cheng'
(15, 10, null, 'I sell carts and also build websites.', 45, 78, '', '2016-12-02 16:02:29.944', '2017-04-05 11:11:25.573', '', '', 'b', '', 0,'7c242316-2b35-11e7-bde0-f40f24220798'),
(16, 10, null, 'I work at transport company and sell ex-goverment cars.', 72, 89, 'From $10,000', '2017-04-27 02:54:21.933', '2017-04-27 02:54:21.933', 'Paypal only', 'Cheap ex-Government cars', 's', '', 0,'7c24235c-2b35-11e7-bde0-f40f24220798'),
(17, 10, null, 'I love computer programming. Various kind of application including mobile, web application.', 65, 95, '$45/hr', '2016-08-13 05:23:19.209', '2017-04-13 05:51:00.726', 'Paypal only', 'Computer Programmer', 'p', 'Computer Engineer', 11,'7c2423a2-2b35-11e7-bde0-f40f24220798'),

--'Liu Jin'
(18, 11, null, 'Default customer profile for buying goods and services.',0,0,'','2017-04-26 06:57:00.631367','2017-04-26 06:57:00.631367','','','b','Customer',0,'7c2423de-2b35-11e7-bde0-f40f24220798'),
(19, 11, null, 'I am best musician in town.',0,0,'100','2017-04-26 10:58:41.604459','2017-04-26 10:58:41.604459','','I can play several musical instruments.','p','Musician',29,'7c242424-2b35-11e7-bde0-f40f24220798'),
(20, 11, null, 'I sell high quality musical instrument.',0,0,'from $500','2017-04-26 11:00:02.826232','2017-04-26 11:00:02.826232','','musical instrument.','s','Sell musical instruments',0,'7c24246a-2b35-11e7-bde0-f40f24220798'),

--'Nuri Kevenoglu'
(21, 12, null, 'I''m a project manager/programmer.', 99, 99, '', '2016-12-02 16:02:29.944', '2017-04-05 11:11:25.573', '', '', 'b', 'Customer', 0,'7c2424a6-2b35-11e7-bde0-f40f24220798');


ALTER SEQUENCE profiles_profile_id_seq RESTART WITH 22;

/****** Insert platform and widget data ******/
---------------------------------
TRUNCATE table public.platforms cascade;
INSERT INTO public.platforms 
(platform_id, name, profile_type, widget_access, created, updated) VALUES
('6876f1a2203311e793ae92361f002671', 'Tool Mates', 'TOOLMATES', true, NOW(), NOW());


TRUNCATE table public.widgets cascade;
INSERT INTO public.widgets
(widget_id, type, owner_id, owner_type, created, updated) VALUES
('aa6cd268273411e793ae92361f002671', 'REPUTATION', '6876f1a2-2033-11e7-93ae-92361f002671', 'PLATFORM', NOW(), NOW());

/* Optionally clear data (highlight script below and execute it to remove all test data)
 
TRUNCATE table public.rooms cascade;
ALTER SEQUENCE rooms_room_id_seq RESTART WITH 1;
TRUNCATE table public.messages cascade;
ALTER SEQUENCE messages_message_id_seq RESTART WITH 1;

TRUNCATE table public.bookings cascade;
ALTER SEQUENCE bookings_booking_id_seq RESTART WITH 1;
TRUNCATE table public.booking_history cascade;
ALTER SEQUENCE booking_history_booking_history_id_seq RESTART WITH 1;

TRUNCATE table public.notifications cascade;
ALTER SEQUENCE notifications_notification_id_seq RESTART WITH 1; 
 
TRUNCATE table public.tags cascade;
 
*/



