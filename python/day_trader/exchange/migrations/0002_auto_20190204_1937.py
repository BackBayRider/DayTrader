# Generated by Django 2.1.5 on 2019-02-04 19:37

from django.db import migrations, models


class Migration(migrations.Migration):

    dependencies = [
        ('exchange', '0001_initial'),
    ]

    operations = [
        migrations.AlterField(
            model_name='baselog',
            name='transaction_num',
            field=models.BigIntegerField(),
        ),
    ]
