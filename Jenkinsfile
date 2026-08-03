// NOTE: Deploy and SonarQube stages are intentionally not yet part of
// platform-standard's standardPipeline() — see the TODOs in the
// platform-standard library repo. They will be added here once ArgoCD and
// SonarQube integration is defined for this pipeline. Until then, this
// Jenkinsfile only calls the shared library and must not gain any stages,
// environment{}, when{}, or steps of its own.

@Library('platform-trunk-based-standard') _

standardPipeline(
    serviceName: 'pokemon-cache-service',
    nodeLabel: 'built-in',
    agentWorkspacePattern: 'workspace/${BRANCH_NAME}/src/git.bluebird.id/platform/pokemon-cache-service'
)
