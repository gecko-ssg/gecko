---
title: "Why Gecko? A smaller path to the web"
date: 2026-10-28
description: "The ideas behind Gecko: readable content, familiar templates, and portable output."
template: "post"
layout: "post"
url: "/blog/building-a-static-site-generator/"
ogType: "article"
tags: ["Gecko", "Go", "Static Sites"]
---

# Why Gecko? A smaller path to the web

Gecko starts with a simple idea: a website should be easy to write, easy to
understand, and easy to move. Content is Markdown, layouts are HTML, and the
result is a directory of ordinary files.

## Write with a light touch

Pages and posts live beside the project source. There is no database to manage
and no application server to keep running after a build.

## Make the markup yours

Gecko uses Go's standard HTML templates, so a shared base layout can define the
document shell while page templates provide their own content. Start from the
included theme, then change the markup and styles to fit your site.

## Publish anywhere

Run `gecko build` and upload the generated `dist/` directory to static hosting.
The output has no Gecko runtime dependency: it is HTML, stylesheets, scripts,
and the assets your pages use.
