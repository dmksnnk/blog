---
date: '2025-04-28T15:30:00+02:00'
title: 'Test smell'
slug: 'test-smell'
description: 'Recognize and address common test smells to improve your tests and code quality.'
cover:
    image: 'images/test-smell.svg'
tags:
  - testing
---

This part is mostly about gut feelings. When you are writing tests or updating the code and something feels wrong, that is a sign that this is a test smell.

- Tests should be your helpers; you should work _with_ them, not against them.

- Tests are your guide for safe refactoring. Tests should enable refactoring, not prevent it. If it seems you are changing internals but need to rewrite the tests, the tests are not testing the behavior, but rather the implementation.

- It is OK to remove tests. Some codebases never remove tests. It can be scary that you are removing guardrails, but some tests can prevent you from making the change or improving the code, instead requiring you to add more workarounds to make the tests go green.

- If your tests have branches, that is a sign you should split them into several tests.

- Tests are hard to read or grasp what they are doing. Tests should be the easiest part of the code to read. Write simple code there; do not be afraid to repeat yourself. DRY (Don't Repeat Yourself) is not always applicable here.

- A flaky test is a sign of a bad test or an actual bug in the code.

- If tests are not easy to run, they will run less often, which will increase the probability of a bug.

- If it's difficult to write tests for a specific piece of code, it is a sign that you need to split the code in more manageable chunks, change the API or having too many dependencies.