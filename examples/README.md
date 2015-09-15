Rfyt REST API Examples
======================

Test Data Generator
-------------------

`logsgen.py` is a pyton test data generation utility. It can be used to produce test files of any size by given template. 
It uses [faker](https://github.com/joke2k/faker) and [pystache](https://github.com/defunkt/pystache) libraries for tempaltes declaration and random data providers.

### Usage
Install Dependencies First:
```
pip install fake-factory
pip install pystache
```

Then see command line arguments: 
```
>python logsgen.py -h
usage: logsgen.py [-h] template count [output]

Generate log files per template.

positional arguments:
  template    template file, use - for <stdin>
  count       number of records to generate
  output      result output file, use - for <stdout> (default)

optional arguments:
  -h, --help  show this help message and exit
```

### Template Spec

See [Mustache](http://mustache.github.io/mustache.5.html) spec for syntax details.

```
See [templates](templates/) for examples.
```


Below is a list of supported random values providers:

```

### faker.providers.address

	fake.random_digit_not_null()                                                                # 8
	fake.building_number()                                                                      # 988
	fake.street_address()                                                                       # 99897 Jessica Pike
	fake.zipcode()                                                                              # 92893
	fake.postalcode_plus4()                                                                     # 53765-0056
	fake.random_digit_not_null_or_empty()                                                       # 3
	fake.city_prefix()                                                                          # New
	fake.military_ship()                                                                        # USS
	fake.country_code()                                                                         # GD
	fake.random_element(elements=('a', 'b', 'b'))                                               # b
	fake.city()                                                                                 # North Maxie
	fake.numerify(text="###")                                                                   # 334
	fake.zipcode_plus4()                                                                        # 63970-1538
	fake.state_abbr()                                                                           # AS
	fake.latitude()                                                                             # 14.7775115
	fake.street_suffix()                                                                        # Spur
	fake.bothify(text="## ??")                                                                  # 86 li
	fake.random_number(digits=None)                                                             # 16130
	fake.city_suffix()                                                                          # view
	fake.random_int(min=0, max=9999)                                                            # 9031
	fake.random_digit_or_empty()                                                                # 9
	fake.military_dpo()                                                                         # Unit 9950 Box 8030
	fake.country()                                                                              # India
	fake.secondary_address()                                                                    # Suite 737
	fake.geo_coordinate(center=None, radius=0.001)                                              # 85.755064
	fake.postalcode()                                                                           # 86331
	fake.address()                                                                              # 8819 Schmeler Track Suite 869
	                                                                                              New Armin, VI 08182-9783
	fake.street_name()                                                                          # Manson Land
	fake.state()                                                                                # Utah
	fake.military_state()                                                                       # AP
	fake.randomize_nb_elements(number=10, le=False, ge=False)                                   # 12
	fake.longitude()                                                                            # -84.700086
	fake.random_letter()                                                                        # Z
	fake.random_digit()                                                                         # 8
	fake.lexify(text="????")                                                                    # BWSv
	fake.postcode()                                                                             # 64810-5104
	fake.military_apo()                                                                         # PSC 4229, Box 6646

### faker.providers.barcode

	fake.ean(length=13)                                                                         # 6049321724732
	fake.ean13()                                                                                # 7057773020349
	fake.ean8()                                                                                 # 05205905

### faker.providers.color

	fake.rgb_css_color()                                                                        # rgb(197,84,17)
	fake.color_name()                                                                           # Ivory
	fake.rgb_color_list()                                                                       # (106, 193, 177)
	fake.rgb_color()                                                                            # 33,249,184
	fake.safe_hex_color()                                                                       # #555500
	fake.safe_color_name()                                                                      # white
	fake.hex_color()                                                                            # #9fc971

### faker.providers.company

	fake.company()                                                                              # Kris PLC
	fake.company_suffix()                                                                       # Group
	fake.catch_phrase()                                                                         # Cross-group zero-defect complexity
	fake.bs()                                                                                   # harness dot-com solutions

### faker.providers.credit_card

	fake.credit_card_security_code(card_type=None)                                              # 688
	fake.credit_card_full(card_type=None, validate=False, max_check=10)                         # 
	                                                                                              JCB 16 digit
	                                                                                              Tressie Schimmel
	                                                                                              3337243334121237  07/17
	                                                                                              CVC: 886
	fake.credit_card_expire(start="now", end="+10y", date_format="%m/%y")                       # 05/18
	fake.credit_card_number(card_type=None, validate=False, max_check=10)                       # 6011810692349067
	fake.credit_card_provider(card_type=None)                                                   # American Express

### faker.providers.currency

	fake.currency_code()                                                                        # AOA

### faker.providers.date_time

	fake.date_time_ad()                                                                         # 1249-05-01 10:51:14
	fake.month()                                                                                # 09
	fake.am_pm()                                                                                # AM
	fake.timezone()                                                                             # Pacific/Palau
	fake.iso8601()                                                                              # 1977-08-19T18:57:06
	fake.date_time()                                                                            # 2004-07-09 22:39:12
	fake.month_name()                                                                           # March
	fake.date_time_this_year(before_now=True, after_now=False)                                  # 2015-03-06 22:37:12
	fake.unix_time()                                                                            # 1286356744
	fake.day_of_week()                                                                          # Wednesday
	fake.day_of_month()                                                                         # 25
	fake.time(pattern="%H:%M:%S")                                                               # 16:28:48
	fake.date_time_between(start_date="-30y", end_date="now")                                   # 2001-08-31 11:41:21
	fake.date_time_this_month(before_now=True, after_now=False)                                 # 2015-09-01 14:11:33
	fake.year()                                                                                 # 1989
	fake.date_time_between_dates(datetime_start=None, datetime_end=None)                        # 2015-09-15 12:18:05
	fake.date_time_this_century(before_now=True, after_now=False)                               # 2011-02-01 00:29:43
	fake.date_time_this_decade(before_now=True, after_now=False)                                # 2013-01-06 04:50:07
	fake.century()                                                                              # XII
	fake.date(pattern="%Y-%m-%d")                                                               # 1970-11-17
	fake.time_delta()                                                                           # 8485 days, 18:19:17

### faker.providers.file

	fake.mime_type(category=None)                                                               # video/mpeg
	fake.file_name(category=None, extension=None)                                               # ut.wav
	fake.file_extension(category=None)                                                          # flac

### faker.providers.internet

	fake.ipv4()                                                                                 # 195.134.73.236
	fake.url()                                                                                  # http://www.littel.com/
	fake.company_email()                                                                        # doyle.carmella@ernser.com
	fake.uri()                                                                                  # http://www.lemke.com/
	fake.domain_word(*args, **kwargs)                                                           # schambergerbecker
	fake.image_url(width=None, height=None)                                                     # http://placekitten.com/988/860
	fake.tld()                                                                                  # com
	fake.free_email()                                                                           # theadore90@hotmail.com
	fake.slug(*args, **kwargs)                                                                  # itaque-dolorem-nam
	fake.free_email_domain()                                                                    # yahoo.com
	fake.domain_name()                                                                          # rohan.info
	fake.uri_extension()                                                                        # .htm
	fake.ipv6()                                                                                 # d06f:d2a5:0b3f:1847:4615:d17c:f21c:dca7
	fake.safe_email()                                                                           # welch.elda@example.com
	fake.user_name(*args, **kwargs)                                                             # ikerluke
	fake.uri_path(deep=None)                                                                    # category/tags/tags
	fake.email()                                                                                # price.shari@champlin.com
	fake.uri_page()                                                                             # faq
	fake.mac_address()                                                                          # 19:84:fc:46:d7:e4

### faker.providers.job

	fake.job()                                                                                  # Scientist, forensic

### faker.providers.lorem

	fake.text(max_nb_chars=200)                                                                 # Qui autem velit aut explicabo voluptates quam similique. E
	                                                                                              t enim corrupti tenetur error. Dolorem maxime quos quia.
	fake.sentence(nb_words=6, variable_nb_words=True)                                           # Et et impedit rerum.
	fake.word()                                                                                 # est
	fake.paragraphs(nb=3)                                                                       # [u'Aliquid omnis placeat eum hic incidunt qui. Nam nihil c
	                                                                                              ulpa eos ipsum consectetur et. Quia aut omnis consequuntur
	                                                                                               qui. In laborum quo nostrum dolorem laborum veniam cumque
	                                                                                              .', u'Ex modi aspernatur totam voluptas consequatur qui en
	                                                                                              im. Expedita est iste fuga omnis excepturi fugit voluptate
	                                                                                              m neque. Dolor molestiae enim sunt quo veniam iure. Conseq
	                                                                                              uatur et tenetur inventore sed velit.', u'Quis officiis cu
	                                                                                              mque voluptas inventore animi ipsam quibusdam. Distinctio 
	                                                                                              in quaerat vitae vero aut voluptatem et. Consequatur eum a
	                                                                                              ut beatae voluptatem voluptates ea nam laboriosam.']
	fake.words(nb=3)                                                                            # [u'ratione', u'sit', u'dignissimos']
	fake.paragraph(nb_sentences=3, variable_nb_sentences=True)                                  # Corporis qui impedit error aperiam tempora maxime et digni
	                                                                                              ssimos. Ea quis repudiandae nobis itaque nulla necessitati
	                                                                                              bus velit nesciunt. Eligendi maiores adipisci exercitation
	                                                                                              em qui aperiam praesentium velit eligendi.
	fake.sentences(nb=3)                                                                        # [u'Nihil laborum quis maiores temporibus.', u'Reiciendis s
	                                                                                              int aut dolorum quaerat assumenda animi placeat.', u'Conse
	                                                                                              quatur molestiae numquam atque maiores est beatae est aspe
	                                                                                              riores.']

### faker.providers.misc

	fake.password(length=10, special_chars=True, digits=True, upper_case=True, lower_case=True) # XaTk%ZUmi$
	fake.locale()                                                                               # pt_BI
	fake.md5(raw_output=False)                                                                  # 176ce20a96a33dd07996ba8cadedd373
	fake.sha1(raw_output=False)                                                                 # 8566719357523fd711184e3697762beffdbef3a4
	fake.null_boolean()                                                                         # False
	fake.sha256(raw_output=False)                                                               # 8c937e3a88a3faea4274a6bec1a819be886c6db57d52914bd7ff9f91df
	                                                                                              e7861e
	fake.uuid4()                                                                                # a56d47da-3911-4812-92ec-981ddbbdf362
	fake.language_code()                                                                        # fr
	fake.boolean(chance_of_getting_true=50)                                                     # True

### faker.providers.person

	fake.last_name_male()                                                                       # Corkery
	fake.name_female()                                                                          # Clair Purdy
	fake.prefix_male()                                                                          # Mr.
	fake.prefix()                                                                               # Ms.
	fake.name()                                                                                 # Dr. Tella Schneider
	fake.suffix_female()                                                                        # DDS
	fake.name_male()                                                                            # Coty Bogisich
	fake.first_name()                                                                           # Olivine
	fake.suffix_male()                                                                          # DDS
	fake.suffix()                                                                               # MD
	fake.first_name_male()                                                                      # Reynolds
	fake.first_name_female()                                                                    # Luda
	fake.last_name_female()                                                                     # Reichel
	fake.last_name()                                                                            # Doyle
	fake.prefix_female()                                                                        # Ms.

### faker.providers.phone_number

	fake.phone_number()                                                                         # 694-089-5577x128

### faker.providers.profile

	fake.simple_profile()                                                                       # {'username': u'hbartoletti', 'name': u'Lollie Armstrong', 
	                                                                                              'birthdate': '2005-02-02', 'sex': 'F', 'address': u'96985 
	                                                                                              Shaniqua Highway\nWest Emanuel, TN 44507', 'mail': u'donne
	                                                                                              lly.zoe@hotmail.com'}
	fake.profile(fields=None)                                                                   # {'website': [u'http://klocko.com/'], 'username': u'corry.j
	                                                                                              erde', 'name': u'Madlyn Dietrich', 'blood_group': 'B+', 'r
	                                                                                              esidence': u'758 Boehm Ferry\nLake Nylah, GU 38551-8388', 
	                                                                                              'company': u'Kihn Inc', 'address': u'6182 Hills Rest\nNort
	                                                                                              h Florianberg, NM 75008', 'birthdate': '1985-04-23', 'sex'
	                                                                                              : 'M', 'job': 'Administrator, education', 'ssn': u'333-82-
	                                                                                              0310', 'current_location': (Decimal('-82.8170995'), Decima
	                                                                                              l('-56.481924')), 'mail': u'koch.esta@gmail.com'}

```
