# Polvo Service

## Cài đặt

> Các command dưới này để được gọi trong Cloud Shell.

1. In Cloud Shell, create the Cloud Storage bucket:

```
PROJECT_ID=$(gcloud config get-value project)
gsutil mb gs://tfstate__${PROJECT_ID}
```

2. Enable Object Versioning to keep the history of your deployments:

```
gsutil versioning set on gs://tfstate__${PROJECT_ID}
```

3. Cấp quyền cho account build

```
CLOUDBUILD_SA="$(gcloud projects describe $PROJECT_ID --format 'value(projectNumber)')@cloudbuild.gserviceaccount.com"
gcloud projects add-iam-policy-binding $PROJECT_ID --member serviceAccount:$CLOUDBUILD_SA --role roles/editor
gcloud projects add-iam-policy-binding $PROJECT_ID --member serviceAccount:$CLOUDBUILD_SA --role roles/run.admin
```

4. Tạo build trigger với substitution variables:

```
_ENV: main
```

`_ENV` này là môi trường đang build, nó sẽ quyết định sự khác biệt về resource.


## Common Use Query

### Delete all versions that do not have package

```
upsert {
  query {
     var(func: eq(dgraph.type, "Version")) @filter(NOT has(~versions)){
     	versionUid as uid
    }
 }
      
  mutation {
    delete {
      uid(versionUid) * * .
    }
  }
}
```

### Detach version

```
upsert {
  query {
     var(func: eq(dgraph.type, "Package")) @filter(eq(name,"sidebar")){
     	packageUid as uid
    }
 }
      
  mutation {
    delete {
      uid(packageUid) <versions>  <0x6d> .
    }
  }
}
```
