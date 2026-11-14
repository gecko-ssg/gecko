<img width="1500" height="500" alt="GECKO" src="https://github.com/user-attachments/assets/1a584df1-936c-4966-a469-6f968b78777b" />

# Gecko

Gecko is a static site generator written in Go. It turns Markdown content and
Go HTML templates into a static website.

## Install

Install the latest release with Go 1.27.3 or newer:

```sh
go install github.com/gecko-ssg/gecko@latest
```

Go places the `gecko` executable in `$GOBIN`, or `$GOPATH/bin` when `GOBIN` is
unset. Make sure that directory is on your `PATH`.

Or download a compiled binary from the [Releases](https://github.com/gecko-ssg/gecko/releases/)
page and add it to your `PATH`.

## Create a site

```sh
gecko init my-site
cd my-site
gecko dev
```

The default `blog` template includes a home page, blog listing, and sample post.
Use `gecko init my-site --template starter` for the same portfolio-style home
page without blog content or post layouts.

`gecko dev` builds the site and starts the development server. When you are
ready to publish, run `gecko build`; Gecko writes the generated website to
`dist/`. Use `gecko preview` to serve a previously built site and `gecko clean`
to remove generated build output.

## Project structure

A site created with `gecko init` has this structure:

```text
my-site/
├── site.yaml                  # Site-wide settings
├── content/
│   ├── index.md               # Home page, portfolio data, and Markdown content
│   ├── blog.md                # Blog listing page
│   └── posts/                 # Blog posts (configured by site.postsDir)
├── layouts/
│   ├── base.html              # Shared HTML document, navigation, and footer
│   └── post.html              # Layout for an individual post
├── templates/
│   ├── index.html             # Home page markup
│   ├── blog.html              # Blog listing markup
│   └── page.html              # Standard page markup
├── assets/
│   ├── css/style-dark.css     # Default stylesheet, copied to dist/assets/
│   └── example.img            # Sample homepage image
├── static/
│   └── favicon.ico            # Static files, copied to the dist/ root
└── dist/                      # Generated site; recreated by gecko build
```

`site.yaml` holds global settings and shared asset links such as the title,
description, author, language, site URL, base URL, stylesheet, favicon, SEO
metadata, theme color, and `postsDir`. Content files are
Markdown with YAML front matter between `---` lines. The `template` field
selects a file in `templates/`; posts use the post layout in `layouts/`.
`layouts/base.html` provides the shared page shell, while the selected template
provides the page body. In the `blog` template, homepage sections are
defined in `templates/index.html`; both homepage templates render `.Content`
from `content/index.md`. Edit `assets/` for files that belong under
`dist/assets/` and `static/` for files copied unchanged to the published site
root.

Avoid editing `dist/` directly: `gecko build` recreates it from the source
directories.

### Change page structure

Edit the HTML template files to change what each page contains:

- `layouts/base.html` defines the document shell shared by the site, including
  the `<head>`, navigation, main area, and footer.
- `templates/index.html` defines the homepage sections.
- `templates/blog.html` defines the posts listing.
- `templates/page.html` defines standard content pages.
- `layouts/post.html` defines individual posts.

Each page template defines a `content` block and renders it inside the `base`
layout. Change the shared navigation or footer in `layouts/base.html`; change a
page's sections or markup in its template. To use a different template for a
Markdown page, set its front matter `template` value to the filename without
`.html` (for example, `template: about` selects `templates/about.html`). The
special `post` template uses `layouts/post.html`.

### Configure portfolio content

The `blog` and `starter` templates read the profile, project cards, and
experience timeline from the home page front matter in `content/index.md`.
Keep this data with the page that displays it:

```yaml
---
title: "Home"
description: "A personal portfolio."
template: "index"
profile:
  name: "Your Name"
  headline: "Software Engineer"
  bio: "A short introduction about yourself."
  image: "/assets/example.img"
  email: "hello@example.com"
  socials:
    - name: GitHub
      url: "https://github.com/username"
    - name: LinkedIn
      url: "https://linkedin.com/in/username"
    - name: X
      url: "https://x.com/username"

projects:
  - name: "Project One"
    description: "A short project description."
    url: "https://github.com/username/project-one"
    technologies: [Go, TypeScript]

experience:
  - company: "Company Name"
    role: "Software Engineer"
    period: "2025 - Present"
    description: "A short description of your work."
---
```

Write the Markdown body after the closing `---`. For example:

```md
# About

This supports **bold**, *italic*, `inline code`, lists, links, images, tables,
task lists, and fenced code blocks.

![Example image](assets/example.img)
```

Put profile and project images under `assets/` and use paths such as
`/assets/example.img` in front matter. The template includes that example
image. The Markdown body in `content/index.md` demonstrates formatting, lists,
links, images, tables, task lists, and code blocks.

## Add a page

From the site directory, run:

```sh
gecko new page About
```

This creates `content/about.md` with front matter and a Markdown heading. Edit
its title, optional description, and body. The default page template is
`page.html`, and the page is published at `/about/`. Use a matching file in
`templates/` if you want a custom page design, then set its name in the page's
`template` front matter field (without `.html`).

## Add a post

```sh
gecko add post "My First Post"
```

Gecko reads `site.postsDir` from `site.yaml` (default: `content/posts`) and
requires that directory to exist before it creates a Markdown file there. The
`starter` template omits blog content and post layouts; choose the `blog`
template for a ready-to-use blog. Add a description
and write the body; the command supplies the date, URL, and `post` template
metadata. Blog posts are rendered with `layouts/post.html` and listed on the
blog page.

## Customize and publish

- Change global site settings in `site.yaml`.
- Edit `templates/index.html`, `templates/blog.html`, or `templates/page.html`
  to change page markup; edit `layouts/` to change shared structure.
- Add images, fonts, and other referenced files under `assets/` or `static/`.
- Run `gecko build` and publish the contents of `dist/` to your static host.

Share feedback, report issues, or contribute through the
[GitHub repository](https://github.com/gecko-ssg/gecko).
