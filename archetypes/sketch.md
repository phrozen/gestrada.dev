---
date: {{ now.Format "2006-01-02" }}
title: "{{ replace .Name "-" " " | title }}"
author: Guillermo Estrada
description: 
rendering: auto
---
<script>
{{<include "" safe>}}
</script>
