- [Contributors Manual](#contributors-manual)
    * [Code Of Conduct](#code-of-conduct)
    * [Opening Issues](#opening-issues)
        + [Bugs](#bugs)
        + [Enhancements](#enhancements)
        + [Feature Requests](#feature-requests)
    * [Contribution Process](#contribution-process)
        + [Pull Requests](#pull-requests)
        + [Signing your commits](#signing-your-commits)
        + [Unblock your users](#unblock-your-users)
    * [Community](#community)

<small><i><a href='http://ecotrust-canada.github.io/markdown-toc/'>Table of contents generated with
markdown-toc</a></i></small>

# Contributors Manual

```
TL;DR;
Look around the project ...
- open and closed Issues section
- Pull Requests
- The documentation

Many times observing the community in practice is the best description of who we are,
and how we act.
```

## Code Of Conduct

Everyone is welcome to use, experiment with or contribute to our open source codebase.

There are some cultural and behavioural traits that we would especially like to promote in this open community. These
are very closely tied and linked to the IKEA values and principles.

Please read our [Code of Conduct](CODE_OF_CONDUCT.md) carefully.

## Opening Issues

Before opening an issue, please take a look at the existing backlog. It might be the case that the content of your
request is already covered.

If your request is already covered in an existing issue, please don't hesitate to upvote or comment
to highlight your support, agreement, or concerns.

`This should be a space for open and elaborate debate ... `
as long as this adheres to the [Code of Conduct](CODE_OF_CONDUCT.md).

### Bugs

For opening bugs, please open an issue while using the
**[bug](../../issues/new?assignees=&labels=bug&template=bug_report.md&title=)**
issue template. Try to be concrete on the detailed in your description, and provide as much of the below information
as possible. This will make the maintainers and contributors work easier and will result in a faster response and
resolution time.

- Software version
- Program inputs / configuration used
- Operational environment
- Steps to reproduce (is the behavior constant or intermittent ?)
- Links or attachments of logs or additional information.
- Do you have a hunch or idea how to solve it?

Before opening a bug issue it is highly recommended checking for any existing ones that are or could be related. Linking
them together will again facilitate the team with the resolution and investigation.

### Enhancements

For pitching ideas or possible enhancements, please open an issue by using the corresponding
**[enhancement](../../issues/new?assignees=&labels=bug&template=enhancement-proposal.md&title=)**
template.

- Describe the idea and rationale. Which use-cases would benefit from this and why ?
- Provide alternative solutions or balance trade-offs between different options.
- Provide diagrams or any accompanying documentation.
- Express your interest and availability if you would like to contribute to this enhancement.

### Feature Requests

The team is always happy to cover more and more use-cases and develop new features. Keep in mind though, that limited
capacity can play an important role in putting your feature in the roadmap.

In that sense, it s always welcome, if feature requests are accompanied by the interest and motivation to contribute to
the idea.

Please open an issue with the corresponding
**[feature](../../issues/new?assignees=&labels=bug&template=feature_request.md&title=)** template and give
a description&mdash;as detailed as possible&mdash;explaining the request.

- Why **this specific** feature?
- Who would benefit from it?
- Do you have any implementation ideas?

## Contribution Process

The below process is based on the practices and approach presented in
the [github collaboration best practices](https://docs.github.com/en/github/collaborating-with-issues-and-pull-requests)

### Pull Requests

**Please communicate with the team** prior to starting your pull request, whether it be by creating an issue or having
  a forum discussion beforehand.  We do not wish to have architectural discussions on a specific
  implementation; ideally those can be handled in Issue threads prior to implementation.  We highly encourage pull
  requests being directly connected to Issues, so that there is some form of documentation that
  the discussion has taken place.

With that being said, here are some concrete steps you should take when preparing a contribution:

- [Fork the repository](https://docs.github.com/en/github/getting-started-with-github/fork-a-repo)
- [Create a Pull Request](https://docs.github.com/en/github/collaborating-with-issues-and-pull-requests/creating-a-pull-request-from-a-fork)
  from your fork to the upstream repository
- Fill in the details in the pull request template

Remember
to [keep your fork up to sync](https://docs.github.com/en/github/collaborating-with-issues-and-pull-requests/syncing-a-fork)
with the main repository

### Signing your commits

Please make sure that all of
your [commits are signed](https://docs.github.com/en/github/authenticating-to-github/signing-commits).

### Unblock your users

In the case where the pull request approval from the maintainers takes too long, or you need to make a ***quick*** fix
for your end users you should use your fork instead of the upstream repository for your downstream dependencies.

This approach comes at a maintenance cost of often syncs and the danger of expanding differences amongst the forks.
**But**, if velocity and usability are your priorities, this should allow you to manage your dependencies and develop
the project at your own pace.

Although the maintainers take much care into solving issues and are open to different approaches and ideas, we are a
community, and as such there is always the possibility of disagreements or longer debates.

### Merging Pull Requests

The Pull requests&mdash;once approved by the owners&mdash;will be merged with a **squash commit**.

The description for the commit will be the **Pull Request title**. (Please choose your title wisely)

Prefer to keep Pull Requests small and concrete in order to enable maintainers to keep a clean commit history.

## Community

A community has only power through its members and for that communication is a vital part of it.

TODO: mention which ways there are for the community to communicate

- forums
- messaging (slack etc ... )
- mailing lists
- community meetings
- ...
