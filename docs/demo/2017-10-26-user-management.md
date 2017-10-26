# Demo - user management - October 26, 2017

This demo covers `/user` API endpoints to manage users.

Note, these REST API endoints are avaialble only for file-based authentication.
The ryft server configuration file should have `auth-type: file`.


## User roles and password

The users text file has been updated so each user now has list of roles.
The most important role for now is `"admin"`. Only users with `"admin"` role
can create new or delete existing users.

There is no plain text passwords in the users file anymore. The HASH function is
used instead:

```{.yaml}
- username: "admin"
  passhash: "$2a$10$ewLPM0A4RJq3iUkMR7FbGep0KxV.JgbeMNeJx0cha.zPb1bZpGlRe"
  roles: ["admin"]
  home: "/"
- username: "test"
  passhash: "$2a$10$tNa9u2ueO2Q4cKiIXCtvWuikwI8lfST04FK6cqTwG5FITIlt7yaC."
  home: "/test"
```

As a side effect the Basic authentication works bit slower than JWT because
ryft server need to calculate and check password hash each time and this is
time consuming operation. So it is recommended to use JWT authentication if possible.


## Get list of users

To get all users the following command can be used:

```{.sh}
$ curl -s "http://localhost:8765/user" -u admin:admin | jq .
[
  {
    "username": "admin",
    "roles": [ "admin" ],
    "home": "/"
  },
  {
    "username": "test",
    "home": "/test",
    "cluster-tag": "test"
  }
]
```

Note, beacuse we are authenticated as `admin` we can request all users.
The `test` user can ask about itself only:

```{.sh}
$ curl -s "http://localhost:8765/user" -u test:test | jq .
[
  {
    "username": "test",
    "home": "/test",
    "cluster-tag": "test"
  }
]
```

There is `name` query parameter which can be used to request particular users:

```{.sh}
$ curl -s "http://localhost:8765/user?name=test" -u admin:admin | jq .
[
  {
    "username": "test",
    "home": "/test",
    "cluster-tag": "test"
  }
]
```

As we said above the non-admin users cannot get information about others:

```{.sh}
$ curl -s "http://localhost:8765/user?name=admin" -u test:test | jq .
{
  "status": 403,
  "message": "access to \"admin\" denied"
}
```


## Create new user

To create new user the `POST` method should be used.

```{.sh}
$ curl -s "http://localhost:8765/user" -u admin:admin -X POST -d '{"username":"foo", "password":"bar", "home":"/foo/bar"}' | jq .
{
  "username": "foo",
  "home": "/foo/bar"
}
```

At least `username` and `password` should be provided in the request's body.
The optional fields are: `roles`, `home` and `cluster-tag`.
Of course the username should be unique.

Only admins can create new users:

```{.sh}
$ curl -s "http://localhost:8765/user" -u test:test -X POST -d '{"username":"foo", "password":"bar", "home":"/foo/bar"}' | jq .
{
  "status": 403,
  "message": "only admin can create new users"
}
```


## Change user

Some user properties can be changed with `PUT` method. New values should be
provided in the request's body (missed values are not changed):

```{.sh}
$ curl -s "http://localhost:8765/user" -u admin:admin -X PUT -d '{"username":"foo", "home":"/test/foo"}' | jq .
{
  "username": "foo",
  "home": "/test/foo",
}

$ curl -s "http://localhost:8765/user" -u admin:admin -X PUT -d '{"username":"foo", "home":"/foo/bar", "roles":["user"]}' | jq .
{
  "username": "foo",
  "roles": [
    "user"
  ],
  "home": "/foo/bar"
}
```

The non-admin users can change the password only:

```{.sh}
$ curl -s "http://localhost:8765/user" -u test:test -X PUT -d '{"username":"test", "password":"test1"}' | jq .
{
  "username": "test",
  "home": "/test",
  "cluster-tag": "test"
}

$ curl -s "http://localhost:8765/user" -u test:test1 -X PUT -d '{"username":"test", "home":"/"}' | jq .
{
  "status": 403,
  "message": "only admin can change home directories"
}
```


## Delete users

To delete users the `DELETE` method should be used.

There is `name` query parameter which can be used to delete particular users:

```{.sh}
$ curl -s "http://localhost:8765/user?name=foo" -u admin:admin -X DELETE | jq .
[
  {
    "username": "foo",
    "roles": [ "user" ],
    "home": "/foo/bar"
  }
]
```

Only admins can delete users:

```{.sh}
$ curl -s "http://localhost:8765/user?name=foo" -u test:test -X DELETE | jq .
{
  "status": 403,
  "message": "only admin can delete users"
}
```
