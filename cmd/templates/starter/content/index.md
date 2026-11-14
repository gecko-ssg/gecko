---
title: "Home"
description: "Personal portfolio, projects, and experience."
template: "index"
layout: "base"
profile:
  name: "Your Name"
  headline: "Software Engineer"
  bio: "A short introduction about yourself."
  image: "/assets/example.img"
  email: "hello@example.com"
  socials:
    - name: "GitHub"
      url: "https://github.com/username"
    - name: "LinkedIn"
      url: "https://linkedin.com/in/username"
    - name: "X"
      url: "https://x.com/username"
projects:
  - name: "Project One"
    description: "A short project description."
    url: "https://github.com/username/project-one"
    image: "/assets/example.img"
    technologies: [Go, TypeScript]
experience:
  - company: "Company Name"
    role: "Software Engineer"
    period: "2025 - Present"
    description: "A short description of your work."
---

# Markdown examples

This homepage content supports **bold**, *italic*, ~~strikethrough~~, and `inline code`.

> Blockquotes make important notes stand out.

- Unordered lists are supported.
- Links work too: [learn about Gecko](https://github.com/gecko-ssg/gecko).

1. Ordered lists work.
2. As do nested Markdown structures.

![Example image from the assets directory](assets/example.img)

| Feature | Syntax |
| --- | --- |
| Table | GitHub Flavored Markdown |
| Heading anchors | Automatic IDs |

- [x] Front matter metadata
- [ ] Customize this sample

```go
func main() {
	fmt.Println("Hello, Gecko!")
}
```
