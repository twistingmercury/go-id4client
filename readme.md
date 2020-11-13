# go-id4client

This package was created so that I could integrate Go apps with a home-grown OAuth2/OIDC solution build on [IdentityServer](https://identityserver.io/) at my current employer.

There may be solutions already existing, but either:
1. I couldn't find them, meaning my google-fu is very weak.
2. I am just too dumb to understand how to use existing ones, or

Believe me, I'd much rather not have to code my own.  I'm a lazy developer.

If there are better solutions out there that will allow me to use a ID4 as a identity provider, please let me know.

:exclamation: Of course, now that they've changed to [Duende Software](https://duendesoftware.com/), this may all be moot in the next few months to a year.  This is because you have to license the IdentityServer product for production use, which is listed as $12,000/yr for [enterprise use](https://duendesoftware.com/products/identityserver#pricing). :expressionless:

## gin middleware

This package is based upon the [Gin Web Framework](https://github.com/gin-gonic/gin).  To use the the middleware, initialize the id4client with the required info, then simply register it as you would any other gin middleware:
```Go
func() main {
    c := id4client.IdentityConfig{
        BaseURL:        "[ID4 BASE URL]",
        ID:             "[APP ID]",
        Secret:         "[APP SECRET]",
        IntrospectPath: "/connect/instrospect",
        TokenPath:      "/connect/token",
        ServiceName:    "[APP NAME]",
        ServiceVersion: "[APP VERSION]",
        CommitHash:     "[GIT COMMIT HASH", //<- this is optional
    }

    if err := id4client.Initialize(c); err != nil{
        log.Fatal(err.Error())
    }
    router := gin.New()
    // use for all end
    r.Use(id4client.Authenticate())

    // setup other middleware(s)
    // setup gin handlers

    router.Run()
}
```
