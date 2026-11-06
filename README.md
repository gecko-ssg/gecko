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
│   ├── index.md               # Home page front matter and Markdown source
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
│   └── css/style.css          # Stylesheet, copied to dist/assets/
├── static/
│   └── favicon.ico            # Static files, copied to the dist/ root
└── dist/                      # Generated site; recreated by gecko build
```

`site.yaml` holds global settings such as the title, description, author,
language, site URL, base URL, theme color, and `postsDir`. Content files are
Markdown with YAML front matter between `---` lines. The `template` field
selects a file in `templates/`; posts use the post layout in `layouts/`.
`layouts/base.html` provides the shared page shell, while the selected template
provides the page body. In the starter portfolio, the homepage sections are
defined in `templates/index.html`; its Markdown body appears only if that
template renders `.Content`. Edit `assets/` for files that belong under
`dist/assets/` and `static/` for files copied unchanged to the published site
root.

Avoid editing `dist/` directly: `gecko build` recreates it from the source
directories.

### Configure portfolio content

The starter portfolio reads its profile, project cards, and experience timeline
from top-level `profile`, `projects`, and `experience` sections in `site.yaml`.
For example:

```yaml
profile:
  name: "Your Name"
  headline: "Software Engineer"
  bio: "A short introduction about yourself."
  image: "/images/profile.png"
  email: "hello@example.com"
  socials:
    github: "https://github.com/username"
    linkedin: "https://linkedin.com/in/username"
    x: "https://x.com/username"

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
```

Put profile and project images under `static/` or `assets/`, then use their
published paths in the configuration. Gecko passes these values to the
homepage template as `.Profile`, `.Projects`, and `.Experience`.

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
gecko new post "My First Post"
```

Gecko reads `site.postsDir` from `site.yaml` (default: `content/posts`) and
creates a Markdown file there. Add the post description and write its body;
the command supplies the date, URL, and `post` template/layout metadata. Posts
are rendered with `layouts/post.html` and appear in the site's blog listing.

## Customize and publish

- Change global site settings in `site.yaml`.
- Edit `templates/index.html`, `templates/blog.html`, or `templates/page.html`
  to change page markup; edit `layouts/` to change shared structure.
- Add images, fonts, and other referenced files under `assets/` or `static/`.
- Run `gecko build` and publish the contents of `dist/` to your static host.

Share feedback, report issues, or contribute through the
[GitHub repository](https://github.com/gecko-ssg/gecko).
