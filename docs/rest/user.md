There are a few REST API endpoints related to user management:

- [GET /user](#list)
- [POST /user](#create)
- [PUT /user](#change)
- [DELETE /user](#delete)

Note, these endpoints are enabled only for `file-based` authentication,
i.e. if `auth-type: file` is in server's configuration file.

Each endpoint is protected and user should provide valid credentials.
See [authentication](../auth.md) for more details.

For most endpoints the authenticated user should have `"admin"` role.
If there is no `"admin"` role the user can change only its password.


# List

The `GET /user` endpoint is used to get list of users.

The list of supported query parameters are the following (check detailed description below):

| Parameter     | Type    | Description |
| ------------- | ------- | ----------- |
| `name`        | string  | [The name of user](#get-user-name-parameter). |

As a result the list of users and its properties are provided.


### GET user `name` parameter

If authenticated user has `"admin"` role then any user with specified name
can be requested. If no name provided then all users will be reported.
Multiple names can be provided as a list `name=foo&name=bar`.

If authenticated user has no `"admin"` role then only that user can be requested.


# Create

The `POST /user` endpoint is used to create a new user.

There is no query parameters but the request body should contain the
following JSON object:

```{.json}
{
  "username":"foo",
  "password":"bar",
  "roles": [ "user" ],
  "home":"/foo"
}
```

Only authenticated user who has `"admin"` role can create new users!


# Change

The `PUT /user` endpoint is used to change some user properties.

There is no query parameters but the request body should contain the
following JSON object:

```{.json}
{
  "username":"foo",
  "password":"bar",
  "roles": [ "user" ],
  "home":"/foo"
}
```

Note, the request fields are optional, it can contains only fields that should be changed.

Only authenticated user who has `"admin"` role can change all properties!
If authenticated user has no `"admin"` role then only password can be changed.


# Delete

The `DELETE /user` endpoint is used to delete users.

The list of supported query parameters are the following (check detailed description below):

| Parameter     | Type    | Description |
| ------------- | ------- | ----------- |
| `name`        | string  | [The name of user](#get-user-name-parameter). |

As a result the list of deleted users and its properties are provided.

Only authenticated user who has `"admin"` role can delete users!
