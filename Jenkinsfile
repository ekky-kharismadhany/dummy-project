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
    agentWorkspacePattern: 'workspace/${BRANCH_NAME}/src/git.bluebird.id/platform/pokemon-cache-service',
    devKubeconfig: 'e252c04a-fd06-43fd-a8fc-7a267aad560e',
    devClusterName: 'dev-huawei-cluster-dummy',
    devNamespace: 'develop',
    stgNamespace: 'staging',
    rgrNamespace: 'regress',
    prdKubeconfig: '9fe21996-2042-44c7-8d60-a23bf7598ae6',
    prdClusterName: 'prd-huawei-cluster-dummy',
    prdNamespace: 'prod',
    devHuaweiProject: 'dev_huawei_project_dummy',
    prdHuaweiProject: 'prd_huawei_project_dummy',
)