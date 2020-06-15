# CLI Tool to mass DM followers on  Twitter

For detailed requirements, refer [here](https://github.com/balajis/twitter-export)

In Brief,

* CLI should,
    * accept arguments like Twitter API Key,Auth token, DM Message
    * Download all followers (with profile details)
    * Rank them by Criteria (e.g. Country)
    * Send each follower a DM with provided message (upto daily DM Limit)
    * be easy to use and maintain

* Notes,
    * Due to Daily DM Limit, Follower details will have to be persisted alongside flag indicating if DM has been sent. SQLITE is good candidate here from simplicity perspective.
    * There should be a scheduled job that will send the DM upto daily DM Limit. At the same time, it needs to refetch any new followers and push them in the flow (reconcile).
    * Potentially, this could be extended to other social media providers other than twitter.
    * Milestones,
        * Create code structure
            * Plan is to have separation between CLI & have twitter as go package
        * Accept Arguments and Connect to Twitter
        * Study and complete follower retrieval
        * Ranking of followers
        * Persisting followers
        * Sending DM upto Daily limit
    * Rules, 
        * Use golang's in-built packages as much as possible
        * Every milestone to have associated Unit test cases