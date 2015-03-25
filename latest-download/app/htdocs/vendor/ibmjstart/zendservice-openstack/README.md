
ZendService\OpenStack
=====================

Provides a simple PHP library for the last version of the [OpenStack API](http://docs.openstack.org/api/api-specs.html).
The goal of this component is to simplify the usage of the OpenStack API for PHP developers, providing simple OO interfaces.

Release notes
-------------

This component is still in development, please don't use it in a production environment.
It uses the [ZendService_Api](https://github.com/zendframework/ZendService_Api) component to manage the HTTP API calls to the OpenStack services.

The development state of each sub-components is reported below:

    - BlockStorage [COMPLETED]
    - Compute [COMPLETED]
    - Identity [COMPLETED]
    - ObjectStorage [COMPLETED]
    - Networking [TO DO]
    - Image [TO DO]

We tested the API using the RackSpace cloud services. We would like to test it using TryStack.org (we are waiting fo the
API support). If you are using a different cloud services that support OpenStack please test this component against it.
In order to execute the online test you need to edit the `TestConfiguration.php.dist` file under the tests folder and change the authentication constants with your account information.
You can run the tests executing the following command under the tests folder:

```
phpunit ZendService/OpenStack
```

Installation
------------
You can install using:

```
curl -s https://getcomposer.org/installer | php
php composer.phar install
```


