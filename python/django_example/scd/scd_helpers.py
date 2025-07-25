

from django.db.models import Max, OuterRef, Subquery

def latest_scd_queryset(model, base_queryset=None):
    """
    Returns a queryset for the latest version of each SCD record in the given model.
    """
    base_queryset = base_queryset or model.objects.all()
    subq = (
        model.objects
        .filter(id=OuterRef('id'))
        .values('id')
        .annotate(max_version=Max('version'))
        .values('max_version')
    )
    return base_queryset.filter(version=Subquery(subq))

def create_new_scd_version(model, id, update_fn):
    latest = model.objects.filter(id=id).order_by('-version').first()
    if not latest:
        raise Exception("Not found")
    new_version = model.objects.get(pk=latest.pk)
    new_version.pk = None
    new_version.version += 1
    update_fn(new_version)
    new_version.save()
    return new_version

 