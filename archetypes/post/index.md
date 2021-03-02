---
draft: true
date: {{ now.Format "2006-01-02" }}
title: "{{ replace .Name "-" " " | title }}"
description:
author: Guillermo Estrada
slug: "{{ .Name }}"
images:
- "&q80&fit=crop&w=1200&h=627"
tags:
categories:
series:
---

{{<cover "">}}