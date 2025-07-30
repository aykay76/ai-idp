This repository contains a dump of some of the conversations i've had with AI agents regarding the introduction of AI agents into Platform Engineering to create an AI-powered Internal Developer Platform (IDP).

My thinking is a bit scattered right now and AI responses are typically quite verbose so i'm trying to organise my thoughts here and constrain myself to one, or a few, specific proof-of-concepts to see if this could even be a thing.

The challenge I face right now is that it's difficult to convince AI agents that they should call APIs on my behalf. They are reluctant to store API keys and don't seem able to retrieve bearer tokens (my preferred approach using OIDC) and they certainly won't act on behalf of a human.

So, I need to find a way of integrating AI without a human in the middle who is responsible for copying and pasting code snippets and manifests into repositories. Maybe it's time to break the mould in terms of files in repos and workflows? Maybe the next evolutionary step is from declarative to NLP in a similar vein to the change from imperative to declarative - it needs a mindshift change in thinking.

