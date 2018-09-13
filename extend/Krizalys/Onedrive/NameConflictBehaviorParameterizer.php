<?php

namespace Krizalys\Onedrive;

class NameConflictBehaviorParameterizer
{
    /**
     * Parameterizes a given name conflict behavior.
     *
     * @param array $params               The parameters.
     * @param int   $nameConflictBehavior The name conflict behavior.
     *
     * @return array
     *
     * @throws \Exception Thrown if the name conflict behavior given is not
     *                    supported.
     */
    public function parameterize(array $params, $nameConflictBehavior)
    {
        switch ($nameConflictBehavior) {
            case NameConflictBehavior::FAIL:
                $params['overwrite'] = 'false';
                break;

            case NameConflictBehavior::RENAME:
                $params['overwrite'] = 'ChooseNewName';
                break;

            case NameConflictBehavior::REPLACE:
                $params['overwrite'] = 'true';
                break;

            default:
                throw new \Exception(
                    "Unsupported name conflict behavior: $nameConflictBehavior"
                );
        }

        return $params;
    }
}
