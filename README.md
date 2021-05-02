# Headb

Headb (pronounced head-bee) is a headless NoSQL database built to enable quick prototyping and apps that do not require
complex database systems.

## What is a headless database?
If a headless CMS is a CMS without any dashboard, then a headless database is a cloud database without any management
consoles, everything happens through an HTTP API.

Many applications do not need a complete MongoDB - or other NoSQL databases - hosting solution to work and be performant.
Yet, there are very few solutions to replace a cloud instance of MongoDB with an API for storing data without limitations.
Services like [npoint](https://www.npoint.io/) or [JSONBin](https://jsonbin.io/), but they have their own limitation
that make them less than ideal for production use.

Headb is an attempt to bridge that gap with a simple, performant, and easy to use solution for storing arbitrary JSON
data on the cloud.

## How to contribute
Headb is build with [encore](https://encore.dev), you will need to go through the set-up to be able to contribute to
the codebase.

Once encore is installed, run `encore run` to start the application. The documentation is built as part of the code
and can be accessed through the encore dashboard.

Run the tests (Coming soon) with `encore test`.